package braintree

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
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

//go:embed graphql/*.graphql
var queryFiles embed.FS

// buildSingleObjectMetadataRequest builds a GraphQL introspection request to fetch field metadata.
// See: https://developer.paypal.com/braintree/graphql/guides/#schema-and-types
func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Convert object name to GraphQL type name for introspection
	// e.g., "customers" -> "Customer", "merchantAccounts" -> "MerchantAccount"
	graphqlTypeName := naming.NewSingularString(naming.CapitalizeFirstLetter(objectName)).String()

	// Use introspection query to get field information including enum values and required status.
	// NON_NULL kind indicates a required field, enumValues provides possible values for enum types.
	query := fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
					enumValues { name }
					ofType {
						name
						kind
						enumValues { name }
						ofType {
							name
							kind
							enumValues { name }
						}
					}
				}
			}
		}
	}`, graphqlTypeName)

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
	req.Header.Add(braintreeVersionHeader, braintreeVersion)

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
		return nil, common.ErrMissingExpectedValues
	}

	// Process each field from the introspection result
	for _, field := range metadataResp.Data.Type.Fields {
		valueType, isRequired, enumValues := extractFieldInfo(field.Type)

		fieldMetadata := common.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(valueType),
			ProviderType: valueType,
			Values:       enumValues,
		}

		if isRequired {
			fieldMetadata.IsRequired = &isRequired
		}

		objectMetadata.Fields[field.Name] = fieldMetadata
	}

	return &objectMetadata, nil
}

// extractFieldInfo extracts the type name, required status, and enum values from a GraphQL type.
// It handles wrapped types like NON_NULL and LIST.
func extractFieldInfo(typeInfo TypeInfo) (typeName string, isRequired bool, enumValues []common.FieldValue) {
	// Check if the field is required (NON_NULL wrapper)
	if typeInfo.Kind == KindNonNull {
		isRequired = true

		// Unwrap to get the actual type
		if typeInfo.OfType != nil {
			return extractFromOfType(*typeInfo.OfType, isRequired)
		}
	}

	// Check for enum values at this level
	if typeInfo.Kind == KindEnum && len(typeInfo.EnumValues) > 0 {
		enumValues = make([]common.FieldValue, len(typeInfo.EnumValues))
		for i, ev := range typeInfo.EnumValues {
			enumValues[i] = common.FieldValue{Value: ev.Name, DisplayValue: ev.Name}
		}

		return typeInfo.Name, isRequired, enumValues
	}

	// Get type name
	if typeInfo.Name != "" {
		return typeInfo.Name, isRequired, nil
	}

	// Unwrap if needed
	if typeInfo.OfType != nil {
		typeName, _, enumValues = extractFromOfType(*typeInfo.OfType, isRequired)

		return typeName, isRequired, enumValues
	}

	return "", isRequired, nil
}

// extractFromOfType extracts type info from nested OfType structures.
func extractFromOfType(ofType OfTypeInfo, isRequired bool) (string, bool, []common.FieldValue) {
	// Check for enum values
	if ofType.Kind == KindEnum && len(ofType.EnumValues) > 0 {
		enumValues := make([]common.FieldValue, len(ofType.EnumValues))
		for i, ev := range ofType.EnumValues {
			enumValues[i] = common.FieldValue{Value: ev.Name, DisplayValue: ev.Name}
		}

		return ofType.Name, isRequired, enumValues
	}

	// If this is NON_NULL, mark as required and continue unwrapping
	if ofType.Kind == KindNonNull {
		isRequired = true

		if ofType.OfType != nil {
			return extractFromOfType(*ofType.OfType, isRequired)
		}
	}

	// Return the type name
	if ofType.Name != "" {
		return ofType.Name, isRequired, nil
	}

	// Continue unwrapping
	if ofType.OfType != nil {
		return extractFromOfType(*ofType.OfType, isRequired)
	}

	return "", isRequired, nil
}

func getFieldValueType(field string) common.ValueType {
	if field == "" {
		return ""
	}

	switch strings.ToLower(field) {
	case "float", "amount", "monetaryamount":
		return common.ValueTypeFloat
	case "string", "id", "timestamp", "date":
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

	var after string

	if params.NextPage != "" {
		after = params.NextPage.String()
	}

	var fromDate, toDate string

	// Note: Braintree's GraphQL Search API only supports filtering by createdAt, not updatedAt.
	// This means incremental reads will only capture newly created records, not updates to existing ones.
	if !params.Since.IsZero() {
		fromDate = datautils.Time.FormatRFC3339inUTC(params.Since)
	}

	if !params.Until.IsZero() {
		toDate = datautils.Time.FormatRFC3339inUTC(params.Until)
	}

	// Use PageSize from params if provided, otherwise use default.
	pageSize := defaultPageSize
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	pagination := graphql.PaginationParameter{
		First:    pageSize,
		After:    after,
		FromDate: fromDate,
		ToDate:   toDate,
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add(braintreeVersionHeader, braintreeVersion)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// Check for GraphQL errors first
	if _, ok := resp.Body(); ok {
		errorResp, err := common.UnmarshalJSON[ResponseError](resp)
		if err == nil && errorResp != nil {
			if checkErr := checkErrorInResponse(errorResp); checkErr != nil {
				return nil, checkErr
			}
		}
	}

	// merchantAccounts uses a different query path: viewer.merchant.merchantAccounts
	// All other objects use the standard search path: search.[objectName]
	// Note: merchantAccounts doesn't have a createdAt field in the schema, so no time filtering is applied.
	if params.ObjectName == objectMerchantAccounts {
		return common.ParseResult(
			resp,
			common.MakeRecordsFunc("edges", "data", "viewer", "merchant", params.ObjectName),
			makeNextRecordsURL(params.ObjectName),
			common.MakeMarshaledDataFunc(common.FlattenNestedFields("node")),
			params.Fields,
		)
	}

	return common.ParseResult(
		resp,
		common.MakeRecordsFunc("edges", "data", "search", params.ObjectName),
		makeNextRecordsURL(params.ObjectName),
		common.MakeMarshaledDataFunc(common.FlattenNestedFields("node")),
		params.Fields,
	)
}

//nolint:cyclop
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	graphqlMutationName := getGraphQLMutationName(params)

	// Build GraphQL mutation with input.
	mutation, err := graphql.Operation(queryFiles, "mutation", graphqlMutationName, nil)
	if err != nil {
		return nil, err
	}

	// Prepare request body with mutation & variables.
	requestBody := map[string]any{
		"query": mutation,
		"variables": map[string]any{
			"input": params.RecordData,
		},
	}

	injectRecordId(params, requestBody)

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add(braintreeVersionHeader, braintreeVersion)

	return req, nil
}

// getGraphQLMutationName determines the GraphQL mutation name based on the operation type.
// For paymentMethods, update is detected by presence of billingAddress in input.
func getGraphQLMutationName(params common.WriteParams) string {
	isUpdate := params.RecordId != ""

	// Special case for paymentMethods: detect update by presence of billingAddress in input.
	if params.ObjectName == objectPaymentMethods && !isUpdate {
		if recordData, err := params.GetRecord(); err == nil {
			if _, hasBillingAddress := recordData["billingAddress"]; hasBillingAddress {
				isUpdate = true
			}
		}
	}

	if isUpdate {
		return params.ObjectName + "Update"
	}

	return params.ObjectName + "Create"
}

// injectRecordId adds the RecordId to the appropriate location in the request body.
// Each object type has different requirements for where the ID should be placed.
func injectRecordId(params common.WriteParams, requestBody map[string]any) {
	if params.RecordId == "" {
		return
	}

	vars, ok := requestBody["variables"].(map[string]any)
	if !ok {
		return
	}

	switch params.ObjectName {
	case "customers":
		// For customers updates, the ID is passed as a separate customerId variable.
		vars["customerId"] = params.RecordId
	case "transactions":
		// For transactions (captureTransaction), inject transactionId into the input.
		if input, ok := vars["input"].(map[string]any); ok {
			input["transactionId"] = params.RecordId
		}
	case objectPaymentMethods:
		// For paymentMethods (updateCreditCardBillingAddress), inject paymentMethodId into the input.
		if input, ok := vars["input"].(map[string]any); ok {
			input["paymentMethodId"] = params.RecordId
		}
	}
}

//nolint:cyclop,funlen
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// Check for GraphQL errors.
	errorResp, err := common.UnmarshalJSON[ResponseError](resp)
	if err == nil && errorResp != nil {
		if checkErr := checkErrorInResponse(errorResp); checkErr != nil {
			return nil, checkErr
		}
	}

	graphqlMutationName := getGraphQLMutationName(params)

	jsonQuery := jsonquery.New(body, "data", graphqlMutationName)

	// For paymentMethods, the response field is "paymentMethod" (singular, camelCase).
	// For other objects, use singular form of the object name.
	var responseFieldName string
	if params.ObjectName == objectPaymentMethods {
		responseFieldName = "paymentMethod"
	} else {
		responseFieldName = naming.NewSingularString(params.ObjectName).String()
	}

	objectResponse, err := jsonQuery.ObjectOptional(responseFieldName)
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
