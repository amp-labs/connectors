package procore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	ErrMissingCompanyID = errors.New("company metadata is required for this object")
	ErrInvalidObject    = errors.New("object name cannot be empty")
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {

	fullObjectEndpoint := resolveAPIPath(objectName, c.companyId)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, fullObjectEndpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Procore-Company-Id", c.companyId)

	return req, nil
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

	records, err := extractRecords(response, objectName)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    analyzeValue(value),
			ProviderType: string(analyzeValue(value)),
		}
	}

	return &objectMetadata, nil
}

// extractRecords returns the list of records from a Procore response.
// Procore returns either a bare array or an object with the array under a "data" key.
func extractRecords(response *common.JSONHTTPResponse, objectName string) ([]any, error) {
	responseKey := readResponseKey.Get(objectName)

	if responseKey != "" {
		obj, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil || obj == nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		data, ok := (*obj)[responseKey].([]any)
		if !ok {
			return nil, fmt.Errorf("%w: response object missing array under \"%s\" key", common.ErrMissingExpectedValues, responseKey)
		}

		return data, nil
	}

	arr, err := common.UnmarshalJSON[[]any](response)
	if err == nil {
		return *arr, nil
	}

	return nil, fmt.Errorf("response body is not in an expected format: %w", common.ErrMissingExpectedValues)
}

func analyzeValue(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
