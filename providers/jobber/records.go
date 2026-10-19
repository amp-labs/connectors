package jobber

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// Jobber webhook payloads carry only record IDs, so subscription processing
// hydrates records here. Each subscribable object has a singular GraphQL
// getter (client(id:), job(id:), ...) embedded as query_<singular>.graphql;
// GetRecordsByIds fans out one query per id with capped concurrency.

const maxConcurrentRecordFetch = 4

var (
	// errRecordFetchNotFound wraps common.ErrNotFound so that when the failure is
	// reported to the caller, common.FailureReasonOf classifies it as
	// FailureReasonNotFound. Jobber signals a missing record with a null node
	// rather than an HTTP status, so there is no 404 to classify from.
	errRecordFetchNotFound = fmt.Errorf("jobber: record not found: %w", common.ErrNotFound)
	errRecordFetchFailed   = errors.New("jobber: record fetch failed")
)

// batchReadableObjects lists objects with a singular GraphQL getter, used to
// hydrate webhook events. It is derived from objectTopicRoot so that every
// subscribable object is hydratable by construction; the unit tests assert
// that each entry also has an embedded getter query file.
//
//nolint:gochecknoglobals
var batchReadableObjects = func() datautils.StringSet {
	set := datautils.NewStringSet()

	for obj := range objectTopicRoot {
		set.AddOne(obj.String())
	}

	return set
}()

//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	if !batchReadableObjects.Has(objectName) {
		return nil, fmt.Errorf("%w: %s", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return common.NewBatchReadResult([]common.ReadResultRow{}), nil
	}

	fieldSet := datautils.NewSetFromList(fields)
	singular := getObjectName(objectName)

	// Tasks are keyed by the id's position rather than the id itself, so the
	// output order follows the requested order and a repeated id still gets its
	// own slot.
	tasks := make([]parallelfetch.Task[int, common.ReadResultRow], len(recordIds))

	for i, recordID := range recordIds {
		idx, id := i, recordID

		tasks[idx] = func(ctx context.Context) (int, *common.ReadResultRow, error) {
			row, err := c.fetchSingleRecord(ctx, singular, id, fieldSet)
			if err != nil {
				return idx, nil, fmt.Errorf("fetch %s/%s: %w", objectName, id, err)
			}

			return idx, &row, nil
		}
	}

	// parallelfetch attempts every id and collects each failure against its own
	// task, so one bad id no longer cancels the fetches still in flight.
	result := parallelfetch.Execute(ctx, tasks, maxConcurrentRecordFetch)

	rows := make([]common.ReadResultRow, 0, len(recordIds))
	for i := range recordIds {
		if row, ok := result.Records[i]; ok {
			rows = append(rows, row)
		}
	}

	batch := common.NewBatchReadResult(rows)

	// Each failure is reported against the id that produced it rather than
	// collapsing the batch into one error. Jobber reports a missing record as a
	// GraphQL error or a null node, both of which surface here as a failure for
	// that id alone.
	for idx, err := range result.Errors {
		batch.AddFailure(recordIds[idx], err)
	}

	batch.SortFailuresByRequestOrder(recordIds)

	return batch, nil
}

// fetchSingleRecord executes the singular getter query for one record id.
func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	singularName, recordID string,
	fieldSet datautils.StringSet,
) (common.ReadResultRow, error) {
	query, err := graphql.Operation(queryFiles, "query", singularName, nil)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	requestBody := map[string]any{
		gqlQueryKey: query,
		gqlVariablesKey: map[string]any{
			"id": recordID,
		},
	}

	resp, err := c.JSONHTTPClient().Post(ctx, c.ProviderInfo().BaseURL, requestBody, versionHeader())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	parsed, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	if parsed == nil {
		return common.ReadResultRow{}, errRecordFetchFailed
	}

	body := *parsed

	// GraphQL reports failures with a 200 status; a missing record surfaces
	// either as an "errors" array or as a null record node.
	if errs, ok := body["errors"].([]any); ok && len(errs) > 0 {
		return common.ReadResultRow{}, fmt.Errorf("%w: %s", errRecordFetchFailed, errorMessages(errs))
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		return common.ReadResultRow{}, errRecordFetchFailed
	}

	record, ok := data[singularName].(map[string]any)
	if !ok || record == nil {
		return common.ReadResultRow{}, errRecordFetchNotFound
	}

	return common.ReadResultRow{
		Id:     recordID,
		Fields: pickFields(record, fieldSet),
		Raw:    record,
	}, nil
}

// pickFields filters raw to entries in fieldSet. An empty fieldSet returns a
// shallow clone. Jobber uses camelCase keys while the framework lower-cases
// requested field names, so matching is case-insensitive.
func pickFields(raw map[string]any, fieldSet datautils.StringSet) map[string]any {
	if len(fieldSet) == 0 {
		return maps.Clone(raw)
	}

	out := make(map[string]any, len(fieldSet))

	for k, v := range raw {
		if fieldSet.Has(k) {
			out[k] = v

			continue
		}

		lower := strings.ToLower(k)
		if fieldSet.Has(lower) {
			out[lower] = v
		}
	}

	return out
}

func errorMessages(errs []any) string {
	messages := make([]string, 0, len(errs))

	for _, e := range errs {
		if errMap, ok := e.(map[string]any); ok {
			if msg, ok := errMap["message"].(string); ok {
				messages = append(messages, msg)
			}
		}
	}

	return strings.Join(messages, "; ")
}
