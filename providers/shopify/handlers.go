package shopify

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

const (
	objectCustomers = "customers"
	objectProducts  = "products"
	objectOrders    = "orders"
)

var (
	ErrMutationDataNotFound   = errors.New("no data found for mutation")
	ErrUnsupportedWriteObject = errors.New("unsupported object for write operation")
)

// perPage is the default number of records per page for Shopify GraphQL API.
// Shopify allows up to 250 records per request, but 100 is chosen as a balanced default
// to avoid rate limiting while maintaining reasonable performance.
// See: https://shopify.dev/docs/api/usage/pagination-graphql
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
		common.ExtractRecordsFromPath("nodes", "data", params.ObjectName),
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

	pageSize := perPage
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	variables["first"] = pageSize

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

// buildWriteRequest constructs a GraphQL mutation request for creating or updating objects.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getDiscoveryEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	mutationKey := getMutationKey(params)

	mutation, err := getMutation(mutationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutation for %s: %w", mutationKey, err)
	}

	// Build request body with mutation and variables
	variables, err := buildWriteVariables(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build write variables: %w", err)
	}

	requestBody := map[string]any{
		"query":     mutation,
		"variables": variables,
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

// parseWriteResponse parses the GraphQL mutation response and extracts the created/updated record.
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	if _, ok := response.Body(); !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// Check for GraphQL userErrors in the response
	writeResp, err := common.UnmarshalJSON[WriteResponse](response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse write response: %w", err)
	}

	mutationKey := getMutationKey(params)
	mutationData, exists := writeResp.Data[mutationKey]
	if !exists {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	if userErrors := parseUserErrors(mutationData); len(userErrors) > 0 {
		var errMsgs []string
		for _, ue := range userErrors {
			errMsgs = append(errMsgs, ue.Message)
		}

		return nil, fmt.Errorf("%w: %s", common.ErrBadRequest, strings.Join(errMsgs, "; "))
	}

	// Extract record ID and object data from the response
	recordID, objectData := extractRecordData(mutationData, params.ObjectName)

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     objectData,
	}, nil
}

// getMutationKey returns the mutation name and GraphQL response key.
func getMutationKey(params common.WriteParams) string {
	singular := naming.NewSingularString(params.ObjectName).String()

	if params.ObjectName == objectOrders && isCloseOrder(params) {
		return "orderClose"
	}

	if params.RecordId != "" {
		return singular + "Update"
	}

	return singular + "Create"
}

// getMutation loads the GraphQL mutation from the embedded filesystem.
func getMutation(mutationName string) (string, error) {
	filePath := fmt.Sprintf("graphql/mutation_%s.graphql", mutationName)

	mutationBytes, err := queryFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read mutation file %s: %w", filePath, err)
	}

	return string(mutationBytes), nil
}

// buildWriteVariables constructs the variables for GraphQL mutations.
func buildWriteVariables(params common.WriteParams) (map[string]any, error) {
	switch params.ObjectName {
	case objectCustomers:
		return buildCustomerVariables(params)
	case objectProducts:
		return buildProductVariables(params)
	case objectOrders:
		return buildOrderVariables(params)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedWriteObject, params.ObjectName)
	}
}

// buildCustomerVariables builds variables for customer create/update mutations.
func buildCustomerVariables(params common.WriteParams) (map[string]any, error) {
	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		record["id"] = params.RecordId
	}

	return map[string]any{
		"input": record,
	}, nil
}

// buildProductVariables builds variables for product create/update mutations.
func buildProductVariables(params common.WriteParams) (map[string]any, error) {
	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		record["id"] = params.RecordId

		return map[string]any{
			"input": record,
		}, nil
	}

	return map[string]any{
		"product": record,
	}, nil
}

// buildOrderVariables builds variables for order create/update/close mutations.
func buildOrderVariables(params common.WriteParams) (map[string]any, error) {
	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		if isCloseOrder(params) {
			return map[string]any{
				"input": map[string]any{
					"id": params.RecordId,
				},
			}, nil
		}

		// Update uses $input with id inside
		record["id"] = params.RecordId

		return map[string]any{
			"input": record,
		}, nil
	}

	return map[string]any{
		"order": record,
	}, nil
}

// parseUserErrors extracts userErrors from the mutation response.
func parseUserErrors(mutationData map[string]any) []UserError {
	userErrorsRaw, ok := mutationData["userErrors"].([]any)
	if !ok || len(userErrorsRaw) == 0 {
		return nil
	}

	var userErrors []UserError

	for _, ueRaw := range userErrorsRaw {
		ue, ok := ueRaw.(map[string]any)
		if !ok {
			continue
		}

		message, _ := ue["message"].(string)
		if message != "" {
			userErrors = append(userErrors, UserError{Message: message})
		}
	}

	return userErrors
}

// extractRecordData extracts the record ID and object data from the mutation response.
func extractRecordData(mutationData map[string]any, objectName string) (string, map[string]any) {
	// Get the singular object name for the response key
	singular := naming.NewSingularString(objectName).String()

	obj, ok := mutationData[singular].(map[string]any)
	if !ok {
		return "", nil
	}

	recordID := ""
	if id, ok := obj["id"].(string); ok {
		recordID = id
	}

	return recordID, obj
}

// =====================================================
// Delete Handlers
// =====================================================

func (c *Connector) buildDeleteRequest(
	ctx context.Context,
	params common.DeleteParams,
) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getDiscoveryEndpoint()
	if err != nil {
		return nil, err
	}

	// Get the mutation name for this delete operation
	mutationName := getDeleteMutationName(params)

	mutation, err := getMutation(mutationName)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutation for %s: %w", mutationName, err)
	}

	// Build the variables for the delete mutation
	variables := buildDeleteVariables(params)

	requestBody := map[string]any{
		"query":     mutation,
		"variables": variables,
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

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	response, err := common.UnmarshalJSON[WriteResponse](resp)
	if err != nil {
		return nil, err
	}

	// Get the mutation key for extracting data
	mutationKey := getDeleteMutationName(params)

	mutationData, ok := response.Data[mutationKey]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMutationDataNotFound, mutationKey)
	}

	userErrors := parseUserErrors(mutationData)
	if len(userErrors) > 0 {
		errorMessages := make([]string, len(userErrors))
		for i, ue := range userErrors {
			errorMessages[i] = ue.Message
		}

		return nil, fmt.Errorf("%w: %s", common.ErrBadRequest, strings.Join(errorMessages, "; "))
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}

// getDeleteMutationName determines the mutation name for delete operations.
func getDeleteMutationName(params common.DeleteParams) string {
	singular := naming.NewSingularString(params.ObjectName).String()

	return singular + "Delete"
}

// buildDeleteVariables constructs the variables for delete mutations.
func buildDeleteVariables(params common.DeleteParams) map[string]any {
	// orderDelete mutation expects variable name `orderId` rather than `id`.
	if params.ObjectName == objectOrders {
		return map[string]any{
			"orderId": params.RecordId,
		}
	}

	return map[string]any{
		"id": params.RecordId,
	}
}

// A closed order is one where merchants fulfill or cancel all LineItem
// objects and complete all financial transactions.
// Once closed, the order indicates that no further work is required.
// https://shopify.dev/docs/api/admin-graphql/latest/mutations/orderclose
func isCloseOrder(params common.WriteParams) bool {
	// Only applies if we have a RecordId
	if params.RecordId == "" {
		return false
	}

	data, ok := params.RecordData.(map[string]any)
	if !ok {
		return false
	}

	// Check if status is "closed"
	if status, ok := data["status"].(string); ok && strings.ToLower(status) == "closed" {
		return true
	}

	return false
}
