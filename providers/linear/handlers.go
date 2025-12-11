package linear

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var perPage = 100 //nolint:gochecknoglobals

//go:embed graphql/*.graphql
var queryFS embed.FS

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
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
					  kind
					}
				}
			}
		}
	}`, naming.NewSingularString(naming.CapitalizeFirstLetter(objectName)).String())

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

		if valueType == "" {
			valueType = field.Type.OfType.Name
		}

		objectMetadata.Fields[field.Name] = common.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(valueType),
			ProviderType: valueType,
			ReadOnly:     false,
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
	case "float":
		return common.ValueTypeFloat
	case "string", "id":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	query := getQuery("graphql/"+params.ObjectName+".graphql", params.ObjectName, params.Fields)

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

	return req, nil
}

// fieldsToTemplateData converts a StringSet of field names to a map for template execution.
// This allows conditional inclusion of nested fields in GraphQL queries.
func fieldsToTemplateData(fields datautils.StringSet) map[string]bool {
	data := make(map[string]bool)

	for field := range fields {
		data[field] = true
	}

	return data
}

func getQuery(filepath, queryName string, fields datautils.StringSet) string {
	queryBytes, err := queryFS.ReadFile(filepath)
	if err != nil {
		return ""
	}

	tmpl, err := template.New(queryName).Parse(string(queryBytes))
	if err != nil {
		return ""
	}

	var queryBuf bytes.Buffer

	// Convert fields slice to a map for template execution
	templateData := fieldsToTemplateData(fields)

	err = tmpl.Execute(&queryBuf, templateData)
	if err != nil {
		return ""
	}

	return queryBuf.String()
}

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

// buildGraphQLVariables creates GraphQL variables for filtering.
func buildGraphQLVariables(params common.ReadParams) map[string]any {
	variables := make(map[string]any)

	variables["first"] = perPage

	if !params.Since.IsZero() {
		filter := map[string]any{
			"updatedAt": map[string]any{
				"gte": params.Since.Format(time.RFC3339Nano),
			},
		}
		variables["filter"] = filter
	}

	if !params.Until.IsZero() {
		filter, exists := variables["filter"].(map[string]any)
		if !exists {
			filter = make(map[string]any)
			variables["filter"] = filter
		}

		if updatedAt, exists := filter["updatedAt"].(map[string]any); exists {
			updatedAt["lte"] = params.Until.Format(time.RFC3339Nano)
		} else {
			filter["updatedAt"] = map[string]any{
				"lte": params.Until.Format(time.RFC3339Nano),
			}
		}
	}

	if params.NextPage != "" {
		variables["after"] = params.NextPage
	}

	return variables
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	query := getQuery("graphql/"+params.ObjectName+"-write.graphql", params.ObjectName, nil)

	requestBody := map[string]any{
		"query": query,
		"variables": map[string]any{
			"input": params.RecordData,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	singularObjName := params.ObjectName[:len(params.ObjectName)-1]

	// Construct the response key based on the singular object name
	// For example, if the object name is "issues", the response key will be "issueCreate"
	responseKey := singularObjName + "Create"

	objectResponse, err := jsonquery.New(node, "data", responseKey).ObjectOptional(singularObjName)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(objectResponse).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	response, err := jsonquery.Convertor.ObjectToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     response,
	}, nil
}
