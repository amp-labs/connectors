package copper

import (
	"context"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Custom Field Definitions.
// https://developer.copper.com/custom-fields/general/list-custom-field-definitions.html
type customFieldsRegistry map[int]customFieldResponse

type customFieldsResponse []customFieldResponse

// nolint:tagliatelle
type customFieldResponse struct {
	ID           int                 `json:"id"`
	DisplayName  string              `json:"name"`
	DataType     string              `json:"data_type"`
	AvailableOn  datautils.StringSet `json:"available_on"`
	IsFilterable bool                `json:"is_filterable"`
	Options      []customFieldOption `json:"options,omitempty"`
	ConnectedId  int                 `json:"connected_id,omitempty"`
	Currency     string              `json:"currency,omitempty"`
}

type customFieldOption struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

func (r customFieldsResponse) FilterByObjectName(objectName string) customFieldsResponse {
	filtered := make([]customFieldResponse, 0)

	for _, field := range r {
		if field.BelongsToObject(objectName) {
			filtered = append(filtered, field)
		}
	}

	return filtered
}

func (r customFieldsResponse) CreateRegistry() customFieldsRegistry {
	registry := make(customFieldsRegistry)

	for _, fields := range r {
		registry[fields.ID] = fields
	}

	return registry
}

func (c customFieldResponse) Name() string {
	// In-house format for custom field.
	return "custom_field_" + strings.ToLower(strings.ReplaceAll(c.DisplayName, " ", "_"))
}

func (c customFieldResponse) BelongsToObject(objectName string) bool {
	return c.AvailableOn.Has(
		naming.NewSingularString(objectName).String(),
	)
}

func (c customFieldResponse) getValues() []common.FieldValue {
	fields := make([]common.FieldValue, 0)

	for _, option := range c.Options {
		fields = append(fields, common.FieldValue{
			Value:        strconv.Itoa(option.Id),
			DisplayValue: option.Name,
		})
	}

	if len(fields) == 0 {
		return nil
	}

	return fields
}

func (c *Connector) fetchCustomFields(ctx context.Context) (*customFieldsResponse, error) {
	url, err := c.getCustomFieldsURL()
	if err != nil {
		return nil, err
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String(), c.emailHeader(), applicationHeader)
	if err != nil {
		return nil, err
	}

	customFields, err := common.UnmarshalJSON[customFieldsResponse](res)
	if err != nil {
		return nil, err
	}

	return customFields, nil
}

func (c *Connector) attachReadCustomFields(customFields *customFieldsResponse) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if customFields == nil || len(*customFields) == 0 {
			// No custom fields, no-op, return as is.
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return enhanceObjectsWithCustomFieldNames(node, customFields.CreateRegistry())
	}
}

func enhanceObjectsWithCustomFieldNames(
	node *ajson.Node, fieldsRegistry customFieldsRegistry,
) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.ParseNode[readCustomFieldsResponse](node)
	if err != nil {
		return nil, err
	}

	// Replace identifiers with human-readable field names which were found by making a call to "/model".
	for _, field := range resp.CustomFields {
		if fieldDefinition, ok := fieldsRegistry[field.ID]; ok {
			object[fieldDefinition.Name()] = field.Value
		}
	}

	return object, nil
}

// nolint:tagliatelle
type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"custom_fields"`
}

// nolint:tagliatelle
type readCustomField struct {
	ID    int `json:"custom_field_definition_id"`
	Value any `json:"value"`
}
