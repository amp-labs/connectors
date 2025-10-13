package chargebee

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	apiVersion = "v2"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	if objectNameWithListSuffix.Has(objectName) {
		objectName += "/list"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	// Add limit parameter to get just a single record for sampling
	url.WithQueryParam("limit", "1")

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	res, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if res == nil || len(*res) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	records, ok := (*res)["list"].([]any) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("couldn't convert the list field to an array: %w", common.ErrMissingExpectedValues)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record to a map: %w", common.ErrMissingExpectedValues)
	}

	objectResponseKey := objectResponseField.Get(objectName)

	// Example response structure
	// 	{
	//     "list": [
	//         {
	//             "customer": {...}
	//         },
	//     ]
	// }

	var objectData map[string]any
	if objectRecord, exists := firstRecord[objectResponseKey]; exists {
		objectData, ok = objectRecord.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the %s object to a map: %w", objectName, common.ErrMissingExpectedValues)
		}
	} else {
		// If the object name key doesn't exist, use the record itself
		objectData = firstRecord
	}

	for field, value := range objectData {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "",
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}
