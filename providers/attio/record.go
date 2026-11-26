package attio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// readByIdsChunkSize is the number of ids sent in one list-records query, and the limit of that query.
// Attio returns at most limit records per query (500 by default), so a chunk no larger than the limit
// fits in one page and the offset never has to be advanced.
// Ref: https://docs.attio.com/rest-api/endpoint-reference/records/list-records
const readByIdsChunkSize = 500

// GetRecordsByIds fetches records by their IDs for a given object type.
//
// Ids are queried in chunks of readByIdsChunkSize. A failed chunk does not fail the call: its ids are reported
// in Failures with that error, the rows of the other chunks are kept, and the chunk errors are returned, joined,
// alongside the result. Ids absent from the responses of the chunks that succeeded are reported as not found.
//
// A 404 fails the whole call instead, see objectNotFoundError.
// Ref: https://docs.attio.com/rest-api/endpoint-reference/records/list-records
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	config := common.ReadParams{
		ObjectName:        objectName,
		Fields:            datautils.NewSetFromList(fields),
		AssociatedObjects: associations,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getObjectReadURL(objectName)
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult([]common.ReadResultRow{})

	var (
		chunkErrs    []error
		succeededIds []string
	)

	for chunk := range slices.Chunk(ids, readByIdsChunkSize) {
		rows, err := c.readChunkByIds(ctx, url, config, chunk)
		if err != nil {
			if notFoundErr := objectNotFoundError(err); notFoundErr != nil {
				return nil, notFoundErr
			}

			for _, id := range chunk {
				result.AddFailure(id, err)
			}

			chunkErrs = append(chunkErrs, err)

			continue
		}

		result.Rows = append(result.Rows, rows...)
		succeededIds = append(succeededIds, chunk...)
	}

	result.ReportMissingIds(succeededIds)

	return result, errors.Join(chunkErrs...)
}

// readChunkByIds runs one list-records query for the given ids.
func (c *Connector) readChunkByIds(
	ctx context.Context, url *urlbuilder.URL, config common.ReadParams, ids []string,
) ([]common.ReadResultRow, error) {
	// Attio's list-records query expects a singular "filter" key.
	// Ref: https://docs.attio.com/rest-api/guides/filtering-and-sorting
	payload := map[string]any{
		"filter": map[string]any{
			"record_id": map[string]any{
				"$in": ids,
			},
		},
		"limit": len(ids),
	}

	res, err := c.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	parsed, err := common.ParseResult(res,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(url, config.ObjectName),
		DataMarshaller(res, config.AssociatedObjects),
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return parsed.Data, nil
}

// objectNotFoundError returns err as a bad request when it is a 404, and nil otherwise.
//
// Attio documents the list-records 404 as "The referenced object was not found": the object in the path, not a
// record id. Every chunk queries the same object, so the whole call fails rather than one chunk.
// The server treats a 404 from GetRecordsByIds as the requested records not existing and drops their events,
// so the 404 is not kept in the returned error chain; only its message is.
func objectNotFoundError(err error) error {
	httpErr, ok := errors.AsType[*common.HTTPError](err)
	if !ok || httpErr.Status != http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("%w: %v", common.ErrBadRequest, err) //nolint:errorlint
}
