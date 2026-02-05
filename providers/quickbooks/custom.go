package quickbooks

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultGraphQLBaseURL = "https://qb.api.intuit.com/graphql"

var objectsWithCustomFields = datautils.NewStringSet( //nolint:gochecknoglobals
	"customer",
	"vendor",
	"invoice",
	"salesReceipt",
	"estimate",
	"creditMemo",
	"refundReceipt",
	"purchaseOrder",
	"bill",
)

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLCustomFieldsResponse struct {
	Data struct {
		AppFoundationsCustomFieldDefinitions []customFieldDefinition `json:"appFoundationsCustomFieldDefinitions"`
	} `json:"data"`
	Errors []graphQLError `json:"errors,omitempty"`
}

type customFieldDefinition struct {
	ID         string `json:"id"`   // Maps to REST CustomField.DefinitionId
	Name       string `json:"name"` // Human-readable name
	Type       string `json:"type"` // StringType, NumberType, DateType, ListType
	LegacyIDV2 string `json:"legacyIdV2,omitempty"`
}

type graphQLError struct {
	Message   string `json:"message"`
	Path      []any  `json:"path,omitempty"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

func (c *Connector) getGraphQLBaseURL() string {
	if c.graphQLBaseURL != "" {
		return c.graphQLBaseURL
	}

	return defaultGraphQLBaseURL
}

func (c *Connector) fetchCustomFieldDefinitions(ctx context.Context) ([]customFieldDefinition, error) {
	query := `query {
		appFoundationsCustomFieldDefinitions {
			id
			name
			type
			legacyIdV2
		}
	}`

	jsonResp, err := c.JSONHTTPClient().Post(ctx, c.getGraphQLBaseURL(), graphQLRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("%w: GraphQL request failed: %w", common.ErrResolvingCustomFields, err)
	}

	graphQLResp, err := common.UnmarshalJSON[graphQLCustomFieldsResponse](jsonResp)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode GraphQL response: %w", common.ErrResolvingCustomFields, err)
	}

	if len(graphQLResp.Errors) > 0 {
		errorMessages := make([]string, len(graphQLResp.Errors))
		for i, gqlErr := range graphQLResp.Errors {
			errorMessages[i] = gqlErr.Message
		}

		return nil, fmt.Errorf("%w: GraphQL errors: %v", common.ErrResolvingCustomFields, errorMessages)
	}

	return graphQLResp.Data.AppFoundationsCustomFieldDefinitions, nil
}

func filterCustomFieldsByObject(fields []customFieldDefinition, objectName string) []customFieldDefinition {
	if !objectsWithCustomFields.Has(objectName) {
		return nil
	}

	return fields
}

// getFieldValueType maps QuickBooks custom field type to common.ValueType.
func getFieldValueType(field customFieldDefinition) common.ValueType {
	switch field.Type {
	case "StringType":
		return common.ValueTypeString
	case "NumberType":
		return common.ValueTypeFloat
	case "DateType":
		return common.ValueTypeDateTime
	case "ListType":
		return common.ValueTypeSingleSelect
	default:
		return common.ValueTypeOther
	}
}

func (c *Connector) attachReadCustomFields(objectName string) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if !objectsWithCustomFields.Has(objectName) {
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return enhanceRecordWithCustomFields(node)
	}
}

// enhanceRecordWithCustomFields flattens the CustomField array values to top-level keys.
func enhanceRecordWithCustomFields(node *ajson.Node) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsResponse, err := jsonquery.ParseNode[readCustomFieldsResponse](node)
	if err != nil {
		return nil, err
	}

	for _, field := range customFieldsResponse.CustomFields {
		if field.Name != "" {
			object[field.Name] = extractCustomFieldValue(field)
		}
	}

	return object, nil
}

func extractCustomFieldValue(field readCustomField) any {
	switch field.Type {
	case "StringType":
		return field.StringValue
	case "NumberType":
		return field.NumberValue
	case "DateType":
		return field.DateValue
	default:
		if field.StringValue != "" {
			return field.StringValue
		}

		return nil
	}
}

type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"CustomField"`
}

// nolint:tagliatelle
type readCustomField struct {
	DefinitionId string `json:"DefinitionId"`
	Name         string `json:"Name"`
	Type         string `json:"Type"`
	StringValue  string `json:"StringValue,omitempty"`
	NumberValue  any    `json:"NumberValue,omitempty"`
	DateValue    string `json:"DateValue,omitempty"`
}
