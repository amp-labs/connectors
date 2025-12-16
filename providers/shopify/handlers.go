package shopify

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

var perPage = 100 //nolint:gochecknoglobals

//go:embed graphql/*.graphql
var queryFS embed.FS

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.getDiscoveryEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Use GraphQL introspection to get field information
	// Convert objectName to PascalCase for GraphQL type names (e.g., "products" -> "Product")
	typeName := naming.NewSingularString(naming.CapitalizeFirstLetter(objectName)).String()

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
					  kind
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
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	metadataResp, err := common.UnmarshalJSON[MetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(metadataResp.Data.Type.Fields) == 0 {
		return nil, fmt.Errorf(
			"missing or empty fields for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	// Process each field from the introspection result
	for _, field := range metadataResp.Data.Type.Fields {
		valueType := field.Type.Name

		if valueType == "" && field.Type.OfType != nil {
			valueType = field.Type.OfType.Name
		}

		objectMetadata.Fields[field.Name] = common.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(valueType),
			ProviderType: valueType,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func getFieldValueType(field string) common.ValueType {
	if field == "" {
		return ""
	}

	switch strings.ToLower(field) {
	case "float", "decimal", "money":
		return common.ValueTypeFloat
	case "string", "id", "url", "html", "date", "datetime":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	case "int", "unsignedint64":
		return common.ValueTypeInt
	default:
		return common.ValueTypeOther
	}
}

// buildReadRequest constructs a GraphQL query request for reading objects.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.getDiscoveryEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	query, err := getQuery(params.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get query for object %s: %w", params.ObjectName, err)
	}

	// Create request body with query and variables
	requestBody := map[string]any{
		"query": query,
	}

	variables := buildGraphQLVariables(params)
	if len(variables) > 0 {
		requestBody["variables"] = variables
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

// parseReadResponse parses the GraphQL response and extracts records.
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

// getQuery loads the GraphQL query from the embedded filesystem.
func getQuery(objectName string) (string, error) {
	filePath := fmt.Sprintf("graphql/%s.graphql", objectName)

	queryBytes, err := queryFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read query file %s: %w", filePath, err)
	}

	return string(queryBytes), nil
}

// buildGraphQLVariables creates GraphQL variables for pagination and filtering.
func buildGraphQLVariables(params common.ReadParams) map[string]any {
	variables := make(map[string]any)

	variables["first"] = perPage

	if params.NextPage != "" {
		variables["after"] = params.NextPage.String()
	}

	// Build Shopify search query for date filtering
	// Shopify uses a query string format: "updated_at:>2024-01-01"
	queryParts := []string{}

	if !params.Since.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("updated_at:>='%s'", params.Since.Format(time.RFC3339)))
	}

	if !params.Until.IsZero() {
		queryParts = append(queryParts, fmt.Sprintf("updated_at:<='%s'", params.Until.Format(time.RFC3339)))
	}

	if len(queryParts) > 0 {
		variables["query"] = strings.Join(queryParts, " AND ")
	}

	return variables
}
