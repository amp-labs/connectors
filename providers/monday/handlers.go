package monday

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	// ErrUnsupportedObject is returned when an unsupported object type is requested.
	ErrUnsupportedObject = errors.New("unsupported object")
	// ErrInvalidResponseFormat is returned when the API response format is unexpected.
	ErrInvalidResponseFormat = errors.New("invalid response format")
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Map object names to their GraphQL type names
	typeNameMap := map[string]string{
		"boards": "Board",
		"users":  "User",
	}

	typeName, exists := typeNameMap[objectName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedObject, objectName)
	}

	// Use introspection query to get field information
	query := fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
					ofType {
						name
					}
				}
			}
		}
	}`, typeName)

	// Create the request body as a map
	requestBody := map[string]string{
		"query": query,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	dataMap, isValidData := (*data)["data"].(map[string]any)
	if !isValidData {
		return nil, fmt.Errorf("%w: missing data field", ErrInvalidResponseFormat)
	}

	typeInfo, exists := dataMap["__type"].(map[string]any)
	if !exists {
		return nil, fmt.Errorf(
			"missing __type in response for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	fields, exists := typeInfo["fields"].([]any)
	if !exists || len(fields) == 0 {
		return nil, fmt.Errorf(
			"missing or empty fields for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	// Process each field from the introspection result
	for _, field := range fields {
		fieldMap, ok := field.(map[string]any)
		if !ok {
			continue
		}

		fieldName, ok := fieldMap["name"].(string)
		if !ok {
			continue
		}

		objectMetadata.FieldsMap[fieldName] = fieldName
	}

	return &objectMetadata, nil
}
