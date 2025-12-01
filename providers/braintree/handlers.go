package braintree

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
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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

	// Convert object name to GraphQL type name for introspection
	// e.g., "customers" -> "Customer", "merchantAccounts" -> "MerchantAccount"
	graphqlTypeName := naming.NewSingularString(naming.CapitalizeFirstLetter(objectName)).String()

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

// objectsWithoutNativeTimeFilter lists objects that don't support createdAt filtering in the API.
// For these objects, we apply connector-side time filtering.
var objectsWithoutNativeTimeFilter = map[string]bool{ //nolint:gochecknoglobals
	objectMerchantAccounts: true,
}

// needsConnectorSideFiltering checks if time filtering should be done connector-side.
func needsConnectorSideFiltering(params common.ReadParams) bool {
	// If no time params, no filtering needed.
	if params.Since.IsZero() && params.Until.IsZero() {
		return false
	}

	// Check if this object doesn't support native time filtering.
	return objectsWithoutNativeTimeFilter[params.ObjectName]
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
	if params.ObjectName == objectMerchantAccounts {
		// Check if we need connector-side time filtering.
		if needsConnectorSideFiltering(params) {
			return common.ParseResultFiltered(
				params,
				resp,
				common.MakeRecordsFunc("edges", "data", "viewer", "merchant", params.ObjectName),
				readhelper.MakeTimeFilterFunc(
					readhelper.Unordered,
					readhelper.NewTimeBoundary(),
					"node.createdAt",
					time.RFC3339,
					makeNextRecordsURL(params.ObjectName),
				),
				common.MakeMarshaledDataFunc(common.FlattenNestedFields("node")),
				params.Fields,
			)
		}

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

	graphqlMutationName := params.ObjectName

	// Determine if this is a create or update operation.
	isUpdate := params.RecordId != ""

	// Special case for paymentMethods: detect update by presence of billingAddress in input.
	// The updateCreditCardBillingAddress mutation uses paymentMethodId in the input, not RecordId.
	if params.ObjectName == objectPaymentMethods && !isUpdate {
		if recordData, ok := params.RecordData.(map[string]any); ok {
			if _, hasBillingAddress := recordData["billingAddress"]; hasBillingAddress {
				isUpdate = true
			}
		}
	}

	if isUpdate {
		graphqlMutationName += "Update"
	} else {
		graphqlMutationName += "Create"
	}

	// Build GraphQL mutation with input.
	mutation, err := graphql.Operation(queryFiles, "mutation", graphqlMutationName, nil)
	if err != nil {
		return nil, err
	}

	// Prepare request body with mutation & variables.
	requestBody := map[string]interface{}{
		"query": mutation,
		"variables": map[string]any{
			"input": params.RecordData,
		},
	}

	if params.RecordId != "" {
		vars, ok := requestBody["variables"].(map[string]any)
		if ok {
			// For customers updates, the ID is passed as a separate customerId variable.
			if params.ObjectName == "customers" {
				vars["customerId"] = params.RecordId
			}
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add(braintreeVersionHeader, braintreeVersion)

	return req, nil
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

	graphqlMutationName := params.ObjectName

	// Determine if this is a create or update operation (same logic as buildWriteRequest).
	isUpdate := params.RecordId != ""

	if params.ObjectName == objectPaymentMethods && !isUpdate {
		if recordData, ok := params.RecordData.(map[string]any); ok {
			if _, hasBillingAddress := recordData["billingAddress"]; hasBillingAddress {
				isUpdate = true
			}
		}
	}

	if isUpdate {
		graphqlMutationName += "Update"
	} else {
		graphqlMutationName += "Create"
	}

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
