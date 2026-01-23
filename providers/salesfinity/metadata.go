package salesfinity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"users": "data.users", // Extract nested users array from team response
}, func(key string) string {
	return "data"
})

// Objects that don't accept limit query parameter.
var objectsWithoutLimit = datautils.NewSet( //nolint:gochecknoglobals
	"users", // Uses team endpoint, doesn't support pagination
)

// Map object names to their actual API endpoints.
var objectNameToEndpoint = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"users":     "team",     // users object uses team endpoint
	"call-logs": "call-log", // call-logs uses singular endpoint
}, func(key string) string {
	return key // Default: use object name as endpoint
})

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	endpoint := objectNameToEndpoint.Get(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v1", endpoint)
	if err != nil {
		return nil, err
	}

	// some endpoints don't accept limit parameter
	if !objectsWithoutLimit.Has(objectName) {
		url.WithQueryParam("limit", "1")
	}

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
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	responseField := ObjectNameToResponseField.Get(objectName)

	data, err := extractDataFromResponse(body, responseField)
	if err != nil {
		return nil, err
	}

	populateFieldsFromMap(data, &objectMetadata)

	return &objectMetadata, nil
}

func extractDataFromResponse(body *ajson.Node, responseField string) (map[string]any, error) {
	// Handle direct object (no wrapper)
	if responseField == "" {
		data, err := jsonquery.Convertor.ObjectToMap(body)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		return data, nil
	}

	// Special handling for users: extract from nested array in team response
	if responseField == "data.users" {
		return extractNestedUsers(body)
	}

	// handle wrapped responses ({data: [...]} or {data: {...}})
	jsonQuery := jsonquery.New(body)

	// try array first
	arr, err := jsonQuery.ArrayOptional(responseField)
	if err == nil && arr != nil {
		if len(arr) == 0 {
			return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
		}

		return jsonquery.Convertor.ObjectToMap(arr[0])
	}

	return nil, fmt.Errorf("couldn't convert %s to array or object: %w", responseField, common.ErrMissingExpectedValues)
}

func extractNestedUsers(body *ajson.Node) (map[string]any, error) {
	jsonQuery := jsonquery.New(body)

	parentNode, err := jsonQuery.ObjectOptional("data")
	if err != nil {
		return nil, fmt.Errorf("couldn't find data object: %w", err)
	}

	arr, err := jsonquery.New(parentNode).ArrayOptional("users")
	if err != nil || arr == nil {
		return nil, fmt.Errorf("couldn't find users array: %w", err)
	}

	if len(arr) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	return jsonquery.Convertor.ObjectToMap(arr[0])
}

func populateFieldsFromMap(data map[string]any, objectMetadata *common.ObjectMetadata) {
	for field, value := range data {
		objectMetadata.AddFieldMetadata(field, common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "",
			Values:       nil,
		})
	}
}

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean

	default:
		return common.ValueTypeOther
	}
}
