package outreach

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// readByIdsChunkSize is the number of ids sent in one filter[id] request, and the page[size] of that request.
// Outreach returns 50 records per page unless page[size] is set, and documents no limit on the number of
// filter values or on URL length. Each comma is sent percent-encoded as %2C, so 200 ids of up to 10 digits
// make a filter[id] value of about 2.6 KB, which keeps the URL short. 200 is also below the page[size] of 1000
// that Read uses.
// https://developers.outreach.io/api/making-requests/
const readByIdsChunkSize = 200

// GetRecordsByIds implements BatchRecordReaderConnector for Outreach.
//
// Ids are requested in chunks of readByIdsChunkSize, and every page of each chunk is read. A failed chunk does
// not fail the call: its ids that were not already returned are reported in Failures with that error, the rows
// of the other chunks are kept, and the chunk errors are returned, joined, alongside the result. Ids absent
// from the responses of the chunks that succeeded are reported as not found.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName:        objectName,
		Fields:            datautils.NewSetFromList(fields),
		AssociatedObjects: associations,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult([]common.ReadResultRow{})

	var (
		chunkErrs    []error
		succeededIds []string
	)

	for chunk := range slices.Chunk(ids, readByIdsChunkSize) {
		rows, err := c.readChunkByIds(ctx, config, chunk)
		result.Rows = append(result.Rows, rows...)

		if err != nil {
			reportChunkFailure(result, chunk, rows, err)
			chunkErrs = append(chunkErrs, err)

			continue
		}

		succeededIds = append(succeededIds, chunk...)
	}

	result.ReportMissingIds(succeededIds)

	return result, errors.Join(chunkErrs...)
}

// readChunkByIds reads every page of one filter[id] request. On error it returns the rows of the pages read
// before the failure.
func (c *Connector) readChunkByIds(
	ctx context.Context, config common.ReadParams, ids []string,
) ([]common.ReadResultRow, error) {
	url, err := c.buildReadByIDsURL(config.ObjectName, ids)
	if err != nil {
		return nil, err
	}

	// Sets the query parameter `include` when the request has associated objects.
	if len(config.AssociatedObjects) > 0 {
		url.WithQueryParam(includeQueryParam, strings.Join(config.AssociatedObjects, ","))
	}

	var rows []common.ReadResultRow

	for next := url.String(); next != ""; {
		rsp, err := c.Client.Get(ctx, next)
		if err != nil {
			return rows, err
		}

		included, err := common.UnmarshalJSON[includedObjects](rsp)
		if err != nil {
			return rows, err
		}

		page, err := common.ParseResult(
			rsp,
			getRecords,
			getNextRecordsURL,
			getOutreachDataMarshaller(config, included.Included, common.FlattenNestedFields(attributesKey)),
			config.Fields,
		)
		if err != nil {
			return rows, err
		}

		rows = append(rows, page.Data...)
		next = page.NextPage.String()
	}

	return rows, nil
}

// reportChunkFailure adds a failure with err for each id of the chunk that is not among the rows read
// before the chunk failed.
func reportChunkFailure(result *common.BatchReadResult, chunk []string, rows []common.ReadResultRow, err error) {
	returned := make(map[string]bool, len(rows))
	for _, row := range rows {
		returned[row.Id] = true
	}

	reported := make(map[string]bool, len(chunk))

	for _, id := range chunk {
		if returned[id] || reported[id] {
			continue
		}

		reported[id] = true

		result.AddFailure(id, err)
	}
}

// buildReadByIDsURL constructs a URL for fetching multiple records by their IDs.
// Uses Outreach API's filter syntax where IDs are comma-separated: filter[id]=1,2,3
// This is more efficient than making individual requests for each ID.
func (c *Connector) buildReadByIDsURL(objectName string, ids []string) (*urlbuilder.URL, error) {
	url, err := c.getApiURL(objectName)
	if err != nil {
		return nil, err
	}

	// Join IDs with commas for the filter parameter
	// Outreach API supports: filter[id]=1,2,3
	if len(ids) > 0 {
		url.WithQueryParam("filter[id]", strings.Join(ids, ","))
	}

	// Without page[size] Outreach returns only the first 50 matching records.
	url.WithQueryParam("page[size]", strconv.Itoa(readByIdsChunkSize))

	return url, nil
}
