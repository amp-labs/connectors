package salesforce

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// recordsByIdsQueryMaxIds is the most ids put in one SOQL query. The ids are sent in the query URL,
// so the limit keeps the URL short.
const recordsByIdsQueryMaxIds = 200

// GetRecordsByIds returns records matching identifiers.
//
// The ids are queried with SOQL, at most 200 per query, and each query is read to its last page.
// A failed query does not fail the call: the ids it did not return are reported as failures with its
// error, the rows of the other queries are returned, and the error, unless it is not-found, is
// returned, joined, alongside the result. Ids no query returned are reported as not found.
func (c *Connector) GetRecordsByIds( // nolint:revive
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

	var (
		result    = common.NewBatchReadResult(nil)
		failed    = datautils.NewStringSet()
		queryErrs []error
	)

	for chunk := range slices.Chunk(uniqueInOrder(ids), recordsByIdsQueryMaxIds) {
		rows, err := c.queryRecordsByIds(ctx, config, chunk)
		result.Rows = append(result.Rows, rows...)

		if err == nil {
			continue
		}

		returned := rowIds(rows)

		for _, id := range chunk {
			if !returned.Has(id) {
				failed.AddOne(id)
				result.AddFailure(id, err)
			}
		}

		// A not-found error is reported only per id, so it is not returned as the error.
		if common.FailureReasonOf(err) != common.FailureReasonNotFound {
			queryErrs = append(queryErrs, fmt.Errorf("query of %d ids: %w", len(chunk), err))
		}
	}

	result.ReportMissingIds(slices.DeleteFunc(slices.Clone(ids), failed.Has))

	return result, errors.Join(queryErrs...)
}

// queryRecordsByIds runs one SOQL query for ids and follows nextRecordsUrl until the query is done.
// When a page fails, it returns the rows of the earlier pages with the error.
func (c *Connector) queryRecordsByIds(
	ctx context.Context, config common.ReadParams, ids []string,
) ([]common.ReadResultRow, error) {
	url, err := c.buildReadByIdentifierURL(config, ids)
	if err != nil {
		return nil, err
	}

	var rows []common.ReadResultRow

	for {
		rsp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			return rows, err
		}

		page, err := common.ParseResult(
			rsp,
			core.GetRecords,
			core.GetNextRecordsURL,
			core.GetDataMarshallerForRead(config),
			config.Fields,
		)
		if err != nil {
			return rows, err
		}

		rows = append(rows, page.Data...)

		if page.Done {
			return rows, nil
		}

		// Same as Read: nextRecordsUrl is a path on the instance domain.
		url, err = c.getDomainURL(page.NextPage.String())
		if err != nil {
			return rows, err
		}
	}
}

func (c *Connector) buildReadByIdentifierURL(config common.ReadParams, ids []string) (*urlbuilder.URL, error) {
	// Requesting record identifiers using SOQL query.
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	soql, err := makeSOQL(config, c.GetTimestampColumn(common.ObjectName(config.ObjectName)))
	if err != nil {
		return nil, err
	}

	// WithIDs quotes the ids in place, so it gets a copy.
	url.WithQueryParam("q", soql.WithIDs(slices.Clone(ids)).String())

	return url, nil
}

// uniqueInOrder removes repeated ids, keeping the first occurrence of each.
func uniqueInOrder(ids []string) []string {
	seen := datautils.NewStringSet()
	unique := make([]string, 0, len(ids))

	for _, id := range ids {
		if seen.Has(id) {
			continue
		}

		seen.AddOne(id)

		unique = append(unique, id)
	}

	return unique
}

// rowIds returns the ids of rows.
func rowIds(rows []common.ReadResultRow) datautils.StringSet {
	ids := datautils.NewStringSet()

	for _, row := range rows {
		ids.AddOne(row.Id)
	}

	return ids
}
