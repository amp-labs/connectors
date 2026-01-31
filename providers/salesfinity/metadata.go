package salesfinity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v1", objectName)
	if err != nil {
		return nil, err
	}

	u.WithQueryParam("limit", "1")

	return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

	data, err := extractDataFromResponse(body, objectName)
	if err != nil {
		return nil, err
	}

	populateFieldsFromMap(data, &objectMetadata)

	return &objectMetadata, nil
}

func extractDataFromResponse(body *ajson.Node, responseField string) (map[string]any, error) {
	// Salesfinity API returns {data: [...]}
	jsonQuery := jsonquery.New(body)

	arr, err := jsonQuery.ArrayOptional("data")
	if err == nil && arr != nil {
		if len(arr) == 0 {
			return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
		}

		return jsonquery.Convertor.ObjectToMap(arr[0])
	}

	return nil, fmt.Errorf("couldn't find data array for %s: %w", responseField, common.ErrMissingExpectedValues)
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
