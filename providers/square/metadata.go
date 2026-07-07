package square

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
	cfg, ok := objects[objectName]
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, objectName)
	}

	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, cfg.path)
	if err != nil {
		return nil, err
	}

	// Fetch a single record to infer field metadata from.
	if cfg.supportsLimit {
		u.WithQueryParam("limit", "1")
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	cfg, ok := objects[objectName] //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, objectName)
	}

	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(naming.SeparateUnderscoreWords(objectName)),
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	data, err := sampleFirstRecord(body, cfg.responseKey)
	if err != nil {
		return nil, err
	}

	populateFieldsFromMap(data, &objectMetadata)

	return &objectMetadata, nil
}

// sampleFirstRecord extracts the first record from the array stored under responseKey.
func sampleFirstRecord(body *ajson.Node, responseKey string) (map[string]any, error) {
	arr, err := jsonquery.New(body).ArrayOptional(responseKey)
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
