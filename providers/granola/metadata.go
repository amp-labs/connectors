package granola

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	// The current version is v0 (alpha).
	// See: https://docs.granola.ai/help-center/sharing/integrations/enterprise-api#what’s-the-api-versioning-strategy
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	u.WithQueryParam("page_size", "1")

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

func extractDataFromResponse(body *ajson.Node, objectName string) (map[string]any, error) {
	jsonQuery := jsonquery.New(body)
	arr, err := jsonQuery.ArrayOptional(objectName)
	if err != nil {
		return nil, err
	}
	if len(arr) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	data, err := jsonquery.Convertor.ObjectToMap(arr[0])
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, common.ErrMissingExpectedValues
	}
	return data, nil
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
