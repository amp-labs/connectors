package jobber

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

//go:embed graphql/*.graphql
var queryFiles embed.FS

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
	}`, objectNameMapping.Get(objectName))

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

	req.Header.Add("X-Jobber-Graphql-Version", apiVersion)

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
		DisplayName: naming.CapitalizeFirstLetter(objectName),
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
		return nil, err
	}

	var (
		after string
		first int
	)

	if params.NextPage != "" {
		after = params.NextPage.String()
	}

	first = defaultPageSize

	pagination := graphql.PaginationParameter{
		First: first,
		After: after,
	}

	query, err := graphql.Operation(queryFiles, "query", params.ObjectName, pagination)
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

	req.Header.Add("X-Jobber-Graphql-Version", apiVersion)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("nodes", "data", params.ObjectName),
		makeNextRecordsURL(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	graphqlQueryName := getObjectName(params.ObjectName)

	if params.RecordId != "" {
		graphqlQueryName += "Edit"
	} else {
		graphqlQueryName += "Create"
	}

	// Build GraphQL mutation with input
	mutation, err := graphql.Operation(queryFiles, "mutation", graphqlQueryName, nil)
	if err != nil {
		return nil, err
	}

	// Prepare request body with mutation & variables
	requestBody := map[string]any{
		"query": mutation,
		"variables": map[string]any{
			"input": params.RecordData,
		},
	}

	if params.RecordId != "" {
		vars, ok := requestBody["variables"].(map[string]any)
		if ok {
			vars["id"] = params.RecordId
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Jobber-Graphql-Version", apiVersion)

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	// When the user provides an invalid field while creating an object,
	// the API returns an error under the "errors" field with a 200 status code.
	// EX. {
	// "errors": [
	//     {
	//         "message": "Argument 'name' on InputObject 'ProductsAndServicesInput' is required. Expected type String!",
	//         "locations": [
	//             {
	//                 "line": 2,
	//                 "column": 38
	//             }
	//         ],...
	errorArr, err := jsonquery.New(body).ArrayOptional("errors")
	if err != nil {
		return nil, err
	}

	if err = checkErrorInResponse(errorArr); err != nil {
		return nil, err
	}

	graphqlQueryName := getObjectName(params.ObjectName)

	if params.RecordId != "" {
		graphqlQueryName += "Edit"
	} else {
		graphqlQueryName += "Create"
	}

	jsonQuery := jsonquery.New(body, "data", graphqlQueryName)

	// User errors appear under the "userErrors" field when existing data is provided while creating an object.
	// Ex.{
	// "data": {
	//     "productsAndServicesCreate": {
	//         "productOrService": null,
	//         "userErrors": [
	//             {
	//                 "message": "A product or service already exists with that name",
	//  .......
	errorArr, err = jsonQuery.ArrayOptional("userErrors")
	if err != nil {
		return nil, err
	}

	if err = checkErrorInResponse(errorArr); err != nil {
		return nil, err
	}

	objectResponse, err := jsonQuery.ObjectOptional(writeObjectNodePathMapping.Get(params.ObjectName))
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
		Errors:   nil,
		Data:     response,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	graphqlQueryName := getObjectName(params.ObjectName)

	graphqlQueryName += "Delete"

	// Generate the mutation string by injecting the record ID.
	// Assumes the template uses a key "record_Id" that maps to params.RecordId
	mutation, err := graphql.Operation(queryFiles, "mutation", graphqlQueryName, nil)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		"query": mutation,
		"variables": map[string]string{
			"id": params.RecordId,
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

	req.Header.Add("X-Jobber-Graphql-Version", apiVersion)

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.DeleteResult{
			Success: true,
		}, nil
	}

	objectResponse, err := jsonquery.New(body).ArrayOptional("errors")
	if err != nil {
		return nil, err
	}

	if err = checkErrorInResponse(objectResponse); err != nil {
		return nil, fmt.Errorf("%w: failed to delete record: %d", err, http.StatusNotFound)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
