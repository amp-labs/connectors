package slack

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/slack/internal/mappings"
	"github.com/amp-labs/connectors/providers/slack/internal/webhook"
)

var _ connectors.BatchRecordReaderConnector = (*Connector)(nil)

// maxConcurrentRecordFetch caps how many ".info" requests run at once. Slack has no bulk
// endpoint for singular records, so each id is one request.
const maxConcurrentRecordFetch = 4

// GetRecordsByIds retrieves records for a given object type by their IDs.
// Slack has no bulk endpoints for singular records, so it makes one request per id,
// concurrently, and always fetches every id.
//
// An id that could not be fetched is reported in the result's Failures with its reason, and the
// rows that did resolve are still returned, in requested order. An id Slack says does not exist
// is a not-found failure and does not produce an error; any other failure is also returned,
// joined, as the error.
// nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) (*common.BatchReadResult, error) {
	if len(recordIds) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Messages cannot be fetched by id — Slack has no singular message lookup — but every
	// message event carries the message itself, so callers are told to read the record
	// from the webhook payload instead of fetching or retrying.
	if webhook.ObjectRecordsInline(objectName) {
		return nil, fmt.Errorf("%w: %s", common.ErrRecordsOnlyInline, objectName)
	}

	identifiers := uniqueInOrder(recordIds)

	info, err := mappings.GetReadItemInfo(c.Provider(), objectName)
	if err != nil {
		// Objects that exist for reading but have no ".info" endpoint, and objects absent
		// from the mappings, are reported with the sentinel that marks the fetch as
		// unavailable, matching the other mapping-driven connectors.
		if errors.Is(err, common.ErrObjectNotSupported) || errors.Is(err, common.ErrOperationNotSupportedForObject) {
			return nil, fmt.Errorf("%w: %s", common.ErrGetRecordNotSupportedForObject, objectName)
		}

		return nil, err
	}

	records, fetchErrs := c.batchRead(ctx, info, identifiers)

	marshaler := readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField(info.ResponseIdField))
	uniqueFields := datautils.NewSetFromList(fields).List()

	rows, err := marshaler(records, uniqueFields)
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult(rows)

	return result, result.ReportFetchErrors(identifiers, fetchErrs)
}

// batchRead fetches each record with its own request, since Slack has no batch endpoints.
//
// Slack resources ending with ".info" (e.g., "channels.info") require a query parameter
// containing a single record identifier. The query parameter name varies by object type.
//
// It returns the fetched records in the order of identifiers, and the errors of the ids
// that failed, keyed by their position in identifiers.
func (c *Connector) batchRead(ctx context.Context,
	info mappings.ReadItemInfo,
	identifiers []string,
) ([]map[string]any, map[int]error) {
	resourceName := info.Href
	queryPramName := info.RequestIdField

	tasks := make([]parallelfetch.Task[int, map[string]any], len(identifiers))
	for index, identifier := range identifiers {
		tasks[index] = func(ctx context.Context) (taskID int, data *map[string]any, err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("%w: resourceName %v(%v=%v)", err, resourceName, queryPramName, identifier)
				}
			}()

			// Make an API URL and make a call.
			url, err := urlbuilder.New(c.ProviderInfo().BaseURL, resourceName)
			if err != nil {
				return index, nil, err
			}

			url.WithQueryParam(queryPramName, identifier)

			res, err := c.JSONHTTPClient().Get(ctx, url.String())
			if err != nil {
				return index, nil, err
			}

			// Unwrap the record from the response and convert to map[string]any.
			body, ok := res.Body()
			if !ok {
				return index, nil, common.ErrEmptyJSONHTTPResponse
			}

			node, err := getResponseSingleRecord(body, info.ResponseField)
			if err != nil {
				return index, nil, err
			}

			apiResponse, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return index, nil, err
			}

			return index, &apiResponse, nil
		}
	}

	result := parallelfetch.Execute(ctx, tasks, maxConcurrentRecordFetch)

	records := make([]map[string]any, 0, len(identifiers))

	for index := range identifiers {
		if record, ok := result.Records[index]; ok {
			records = append(records, record)
		}
	}

	return records, result.Errors
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
