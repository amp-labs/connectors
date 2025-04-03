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
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = 2
)

var (
	// ErrUnsupportedObject is returned when an unsupported object type is requested.
	ErrUnsupportedObject = errors.New("unsupported object")
	// ErrInvalidResponseFormat is returned when the API response format is unexpected.
	ErrInvalidResponseFormat = errors.New("invalid response format")
	// ErrUnsupportedObjectName is returned when an unsupported object name is provided.
	ErrUnsupportedObjectName = errors.New("unsupported object name")
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

func getBoardsBaseFields() string {
	return `
		id
		name
		state
		permissions
		items_count
		type
		updated_at
		url
		workspace_id`
}

func getBoardsNestedFields() string {
	return `
		columns {
			id
			title
			type
		}
		groups {
			id
			title
			position
		}
		owner {
			id
			name
		}
		owners {
			id
			name
		}
		subscribers {
			id
			name
		}
		tags {
			id
			name
		}
		team_owners {
			id
			name
		}
		team_subscribers {
			id
			name
		}
		top_group {
			id
			title
		}
		updates {
			id
			body
			created_at
		}
		views {
			id
			name
			type
		}
		workspace {
			id
			name
		}`
}

func getBoardsQuery(page *int, limit *int) string {
	paginationParams := ""
	if page != nil && limit != nil {
		paginationParams = fmt.Sprintf("(limit: %d, page: %d)", *limit, *page)
	}

	return fmt.Sprintf(`query {
		boards%s {
			%s
			%s
		}
	}`, paginationParams, getBoardsBaseFields(), getBoardsNestedFields())
}

func getUsersQuery(page *int, limit *int) string {
	paginationParams := ""
	if page != nil && limit != nil {
		paginationParams = fmt.Sprintf("(limit: %d, page: %d)", *limit, *page)
	}

	return fmt.Sprintf(`query {
		users%s {
			id
			email
			name
			enabled
		}
	}`, paginationParams)
}

func getQueryForObject(objectName string, page *int, limit *int) (string, error) {
	switch objectName {
	case "boards":
		return getBoardsQuery(page, limit), nil
	case "users":
		return getUsersQuery(page, limit), nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedObjectName, objectName)
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion)
	if err != nil {
		return nil, err
	}

	var page *int

	var limit int

	if params.NextPage != "" {
		// Parse the page number from NextPage
		var pageNum int

		_, err := fmt.Sscanf(string(params.NextPage), "%d", &pageNum)
		if err != nil {
			return nil, fmt.Errorf("invalid next page format: %w", err)
		}

		page = &pageNum
		limit = defaultPageSize
	}

	query, err := getQueryForObject(params.ObjectName, page, &limit)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]string{
		"query": query,
	}

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

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	data, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	dataMap, isValidData := (*data)["data"].(map[string]any)
	if !isValidData {
		return nil, fmt.Errorf("%w: missing data field", ErrInvalidResponseFormat)
	}

	rawRecords, exists := dataMap[params.ObjectName]
	if !exists {
		errMsg := "missing expected values for object: " + params.ObjectName

		return nil, fmt.Errorf("%s, error: %w", errMsg, common.ErrMissingExpectedValues)
	}

	records, isValidRecords := rawRecords.([]any)
	if !isValidRecords {
		errMsg := "unexpected type for records for object: " + params.ObjectName

		return nil, fmt.Errorf("%s, error: %w", errMsg, common.ErrMissingExpectedValues)
	}

	return common.ParseResult(
		resp,
		getRecords(params.ObjectName),
		makeNextRecordsURL(params, len(records)),
		common.GetMarshaledData,
		params.Fields,
	)
}

func getRecords(objectName string) func(*ajson.Node) ([]map[string]any, error) {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// First get the data object
		dataNode, err := node.GetKey("data")
		if err != nil {
			return nil, err
		}

		// Then get the array under the object name (e.g., "boards" or "users")
		records, err := jsonquery.New(dataNode).ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}
