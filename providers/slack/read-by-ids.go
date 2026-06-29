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
)

var _ connectors.BatchRecordReaderConnector = (*Connector)(nil)

// GetRecordsByIds retrieves records for a given object type by their IDs.
// It performs batched parallel reads since Slack has no bulk endpoints for singular records.
// nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	objectName string, recordIds []string,
	fields []string, associations []string,
) ([]common.ReadResultRow, error) {
	if len(recordIds) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Ensure identifiers are non-repeating.
	identifiers := datautils.NewSetFromList(recordIds).List()

	batchResult, err := c.batchRead(ctx, objectName, identifiers)
	if err != nil {
		return nil, err
	}

	marshaler := readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id"))
	uniqueFields := datautils.NewSetFromList(fields).List()

	return marshaler(batchResult, uniqueFields)
}

// batchRead performs parallel reads for multiple singular records.
//
// Slack resources ending with ".info" (e.g., "channels.info") require a query parameter
// containing a single record identifier. The query parameter name varies by object type,
// and only one identifier can be sent per URL—there are no batch endpoints.
//
// This method makes multiple parallel requests to fetch each record individually.
func (c *Connector) batchRead(ctx context.Context, objectName string, identifiers []string) ([]map[string]any, error) {
	resourceName := objectName + ".info"

	queryPramName, ok := readSingleRecordResourceNameToQueryParam[resourceName]
	if !ok {
		return nil, fmt.Errorf("%w: object name [%v]", common.ErrOperationNotSupportedForObject, objectName)
	}

	tasks := make([]parallelfetch.Task[string, map[string]any], len(identifiers))
	for index, identifier := range identifiers {
		tasks[index] = func(ctx context.Context) (taskID string, data *map[string]any, err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("%w: resourceName %v(%v=%v)", err, resourceName, queryPramName, identifier)
				}
			}()

			// Make an API URL and make a call.
			url, err := urlbuilder.New(c.ProviderInfo().BaseURL, resourceName)
			if err != nil {
				return identifier, nil, err
			}

			url.WithQueryParam(queryPramName, identifier)

			res, err := c.JSONHTTPClient().Get(ctx, url.String())
			if err != nil {
				return identifier, nil, err
			}

			// Unwrap the record from the response and convert to map[string]any.
			body, ok := res.Body()
			if !ok {
				return identifier, nil, common.ErrEmptyJSONHTTPResponse
			}

			node, err := getResponseSingleRecord(body, resourceName)
			if err != nil {
				return identifier, nil, err
			}

			apiResponse, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return identifier, nil, err
			}

			return identifier, &apiResponse, nil
		}
	}

	result := parallelfetch.Execute(ctx, tasks, -1)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return result.Records.Values(), nil
}
