package microsoft

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

var _ connectors.BatchRecordReaderConnector = (*Connector)(nil)

// GetRecordsByIds scoped reading of records given their ids.
//
// Each id is one request inside Graph's $batch. A failed request does not fail the call: every failed
// id is reported in Failures, and the rows of the ids that resolved are returned. Failures other than
// not-found are also returned, joined, as the error.
// nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) (*common.BatchReadResult, error) {
	if len(recordIds) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Graph rejects a whole $batch request when two of its requests share an id.
	ids := uniqueInOrder(recordIds)

	batchParams, requestIdentifiers, err := c.paramsForBatchRead(objectName, ids)
	if err != nil {
		return nil, err
	}

	batchResponse := batch.Execute[map[string]any](ctx, c.batchStrategy, batchParams)

	marshaler := readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id"))
	uniqueFields := datautils.NewSetFromList(fields).List()

	rows, err := marshaler(batchResponse.GetInOrder(requestIdentifiers), uniqueFields)
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult(rows)

	return result, result.ReportFetchErrors(ids, batchErrorsByPosition(batchResponse, ids, requestIdentifiers))
}

// batchErrorsByPosition keys each failed id's error by its position in ids, as ReportFetchErrors expects.
// An id Graph left out of its response is reported as not found.
func batchErrorsByPosition(
	batchResponse *batch.Result[map[string]any], ids []string, requestIdentifiers []batch.RequestID,
) map[int]error {
	errs := make(map[int]error)

	for index, requestIdentifier := range requestIdentifiers {
		if failure, ok := batchResponse.Errors[requestIdentifier]; ok {
			errs[index] = failure.Data

			continue
		}

		if _, ok := batchResponse.Responses[requestIdentifier]; !ok {
			errs[index] = fmt.Errorf("%w: %s not returned by the batch request", common.ErrNotFound, ids[index])
		}
	}

	return errs
}

// uniqueInOrder removes repeated ids, keeping the first occurrence of each.
func uniqueInOrder(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	unique := make([]string, 0, len(ids))

	for _, id := range ids {
		if seen[id] {
			continue
		}

		seen[id] = true

		unique = append(unique, id)
	}

	return unique
}

func (c *Connector) paramsForBatchRead(
	objectName string, identifiers []string,
) (*batch.Params, []batch.RequestID, error) {
	batchParams := &batch.Params{}

	requestIdentifiers := make([]batch.RequestID, len(identifiers))
	for index, identifier := range identifiers {
		url, err := c.getReadUrl(objectName)
		if err != nil {
			return nil, nil, err
		}

		url.AddPath(identifier)
		requestIdentifier := batch.RequestID(fmt.Sprintf("%v_%v", objectName, identifier))
		requestIdentifiers[index] = requestIdentifier
		batchParams.WithRequest(requestIdentifier, http.MethodGet, url, nil, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, requestIdentifiers, nil
}
