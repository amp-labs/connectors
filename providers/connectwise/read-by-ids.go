package connectwise

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/batch"
	"github.com/amp-labs/connectors/providers/connectwise/internal/reader"
)

// GetRecordsByIds scoped reading of records given their ids.
//
// The ids are split across several requests. A failed request does not fail the call: each of its
// ids is reported in Failures with the request's error, the rows returned by the other requests are
// kept, and the errors of the failed requests are returned, joined, except not-found errors. An id
// that no request returned and no failed request asked for is reported as not found.
// nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) (*common.BatchReadResult, error) {
	if len(recordIds) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Ensure identifiers are non-repeating.
	ids := datautils.NewSetFromList(recordIds).List()

	batchResult, err := batch.Read[map[string]any](ctx, c.Base, objectName, ids)
	if err != nil {
		return nil, err
	}

	marshaler := readhelper.MakeMarshaledDataFuncWithId(
		reader.RecordTransformer(objectName),
		readhelper.IdFieldQuery{Field: "id"},
	)
	uniqueFields := datautils.NewSetFromList(fields).List()

	list, err := jsonquery.Convertor.NodesFromArray(batchResult.Records)
	if err != nil {
		return nil, err
	}

	rows, err := marshaler(list, uniqueFields)
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult(rows)
	failedIds := datautils.NewStringSet()

	// A not-found failure is reported only per id, like a missing id, so the call returns no error for it.
	var errs []error

	for _, failure := range batchResult.Failures {
		for _, id := range failure.Identifiers {
			result.AddFailure(id, failure.Err)
			failedIds.AddOne(id)
		}

		if common.FailureReasonOf(failure.Err) != common.FailureReasonNotFound {
			errs = append(errs, failure.Err)
		}
	}

	// Ids of failed requests already have a failure.
	result.ReportMissingIds(withoutIds(recordIds, failedIds))

	return result, errors.Join(errs...)
}

// withoutIds returns the ids that are not in exclude, keeping their order.
func withoutIds(ids []string, exclude datautils.StringSet) []string {
	kept := make([]string, 0, len(ids))

	for _, id := range ids {
		if !exclude.Has(id) {
			kept = append(kept, id)
		}
	}

	return kept
}
