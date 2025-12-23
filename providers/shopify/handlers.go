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
	objectCustomerAddresses      = "customerAddresses"
	objectCustomerDefaultAddress = "customerDefaultAddress"
	objectProducts               = "products"
	objectProductOptions         = "productOptions"
	compositeIDParts             = 2
)

var ErrMutationDataNotFound = errors.New("no data found for mutation")

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

	mutationName := getMutationName(params)

	mutation, err := getMutation(mutationName)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutation for %s: %w", mutationName, err)
	}

	// Build request body with mutation and variables
	variables := buildWriteVariables(params)

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
	if mutationData, exists := writeResp.Data[mutationKey]; exists {
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

	return &common.WriteResult{
		Success: true,
	}, nil
}

// getMutationName determines the mutation name based on object and operation type.
func getMutationName(params common.WriteParams) string {
	if params.ObjectName == objectCustomerAddresses {
		if params.RecordId != "" {
			return "customerAddressUpdate"
		}

		return "customerAddressCreate"
	}

	// Handle customerDefaultAddress - always an update operation.
	if params.ObjectName == objectCustomerDefaultAddress {
		return "customerUpdateDefaultAddress"
	}

	// Handle productOptions - always a create operation.
	if params.ObjectName == objectProductOptions {
		return "productOptionsCreate"
	}

	// Convert plural object name to singular for mutation name, e.g., "customers" -> "customer"
	singular := naming.NewSingularString(params.ObjectName).String()

	if params.RecordId != "" {
		return singular + "Update"
	}

	return singular + "Create"
}

// getMutationKey returns the GraphQL response key for the mutation.
func getMutationKey(params common.WriteParams) string {
	// Handle customerAddresses as a special case.
	if params.ObjectName == objectCustomerAddresses {
		if params.RecordId != "" {
			return "customerAddressUpdate"
		}

		return "customerAddressCreate"
	}

	// Handle customerDefaultAddress.
	if params.ObjectName == objectCustomerDefaultAddress {
		return "customerUpdateDefaultAddress"
	}

	// Handle productOptions.
	if params.ObjectName == objectProductOptions {
		return "productOptionsCreate"
	}

	singular := naming.NewSingularString(params.ObjectName).String()

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
func buildWriteVariables(params common.WriteParams) map[string]any {
	if params.ObjectName == objectCustomerAddresses {
		return buildCustomerAddressVariables(params)
	}

	// Handle customerDefaultAddress, uses $customerId and $addressId.
	if params.ObjectName == objectCustomerDefaultAddress {
		return buildCustomerDefaultAddressVariables(params)
	}

	// Handle productOptions - uses $productId, $options, and optional $variantStrategy.
	if params.ObjectName == objectProductOptions {
		return buildProductOptionsVariables(params)
	}

	// Handle products - uses $product for create, $input for update.
	if params.ObjectName == objectProducts {
		return buildProductVariables(params)
	}

	variables := map[string]any{
		"input": params.RecordData,
	}

	if params.RecordId != "" {
		injectIDIntoInput(variables, params.RecordId)
	}

	return variables
}

// buildCustomerAddressVariables builds variables for customerAddress mutations.
func buildCustomerAddressVariables(params common.WriteParams) map[string]any {
	recordData, ok := params.RecordData.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	variables := make(map[string]any)

	if customerId, exists := recordData["customerId"]; exists {
		variables["customerId"] = customerId
	}

	if address, exists := recordData["address"]; exists {
		variables["address"] = address
	}

	if setAsDefault, exists := recordData["setAsDefault"]; exists {
		variables["setAsDefault"] = setAsDefault
	}

	if params.RecordId != "" {
		variables["addressId"] = params.RecordId
	}

	return variables
}

// buildCustomerDefaultAddressVariables builds variables for customerUpdateDefaultAddress mutation.
func buildCustomerDefaultAddressVariables(params common.WriteParams) map[string]any {
	recordData, ok := params.RecordData.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	variables := make(map[string]any)

	if customerId, exists := recordData["customerId"]; exists {
		variables["customerId"] = customerId
	}

	if addressId, exists := recordData["addressId"]; exists {
		variables["addressId"] = addressId
	}

	return variables
}

// buildProductVariables builds variables for product create/update mutations.
func buildProductVariables(params common.WriteParams) map[string]any {
	if params.RecordId != "" {
		// Update uses $input with id inside.
		variables := map[string]any{
			"input": params.RecordData,
		}

		injectIDIntoInput(variables, params.RecordId)

		return variables
	}

	// Create uses $product.
	return map[string]any{
		"product": params.RecordData,
	}
}

// buildProductOptionsVariables builds variables for productOptionsCreate mutation.
func buildProductOptionsVariables(params common.WriteParams) map[string]any {
	recordData, ok := params.RecordData.(map[string]any)
	if !ok {
		return map[string]any{}
	}

	variables := make(map[string]any)

	if productId, exists := recordData["productId"]; exists {
		variables["productId"] = productId
	}

	if options, exists := recordData["options"]; exists {
		variables["options"] = options
	}

	if variantStrategy, exists := recordData["variantStrategy"]; exists {
		variables["variantStrategy"] = variantStrategy
	}

	return variables
}

// injectIDIntoInput adds the record ID inside the input for update operations.
func injectIDIntoInput(variables map[string]any, recordID string) {
	input, ok := variables["input"].(map[string]any)
	if !ok {
		return
	}

	input["id"] = recordID
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
	switch objectName {
	case objectCustomerAddresses:
		return extractFromKey(mutationData, "address")
	case objectCustomerDefaultAddress:
		return extractCustomerDefaultAddress(mutationData)
	case objectProductOptions:
		return extractFromKey(mutationData, "product")
	default:
		singular := naming.NewSingularString(objectName).String()

		return extractFromKey(mutationData, singular)
	}
}

// extractFromKey extracts record ID and data from a specific key in the mutation response.
func extractFromKey(mutationData map[string]any, key string) (string, map[string]any) {
	obj, ok := mutationData[key].(map[string]any)
	if !ok {
		return "", nil
	}

	recordID, _ := obj["id"].(string)

	return recordID, obj
}

// extractCustomerDefaultAddress extracts data from customerDefaultAddress mutation response.
func extractCustomerDefaultAddress(mutationData map[string]any) (string, map[string]any) {
	customer, ok := mutationData["customer"].(map[string]any)
	if !ok {
		return "", nil
	}

	defaultAddress, ok := customer["defaultAddress"].(map[string]any)
	if !ok {
		return "", customer
	}

	recordID, _ := defaultAddress["id"].(string)

	return recordID, defaultAddress
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
	if params.ObjectName == objectCustomerAddresses {
		return "customerAddressDelete"
	}

	// Handle productOptions deletion.
	if params.ObjectName == objectProductOptions {
		return "productOptionsDelete"
	}

	// Convert plural object name to singular for mutation name.
	singular := naming.NewSingularString(params.ObjectName).String()

	return singular + "Delete"
}

// buildDeleteVariables constructs the variables for delete mutations.
func buildDeleteVariables(params common.DeleteParams) map[string]any {
	// RecordId format: "customerId|addressId".
	if params.ObjectName == objectCustomerAddresses {
		parts := strings.SplitN(params.RecordId, "|", compositeIDParts)
		if len(parts) == compositeIDParts {
			return map[string]any{
				"customerId": parts[0],
				"addressId":  parts[1],
			}
		}

		return map[string]any{
			"addressId": params.RecordId,
		}
	}

	// RecordId format for productOptions: "productId|optionId1,optionId2,...".
	if params.ObjectName == objectProductOptions {
		parts := strings.SplitN(params.RecordId, "|", compositeIDParts)
		if len(parts) == compositeIDParts {
			optionIDs := strings.Split(parts[1], ",")

			return map[string]any{
				"productId": parts[0],
				"options":   optionIDs,
			}
		}

		return map[string]any{
			"options": []string{params.RecordId},
		}
	}

	return map[string]any{
		"id": params.RecordId,
	}
}
