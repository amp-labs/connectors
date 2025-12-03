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

	pagination := graphql.PaginationParameter{
		First:    defaultPageSize,
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
	if params.ObjectName == "merchantAccounts" {
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
