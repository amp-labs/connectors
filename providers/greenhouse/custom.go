package greenhouse

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// fieldTypeToObjectName maps Greenhouse custom field field_type to connector object names.
// https://harvestdocs.greenhouse.io/reference/get_v3-custom-fields
var fieldTypeToObjectName = map[string]string{ //nolint:gochecknoglobals
	"candidate":      "candidates",
	"application":    "applications",
	"job":            "jobs",
	"opening":        "openings",
	"offer":          "offers",
	"user_attribute": "users",
}

// objectNameToFieldType is the reverse mapping.
var objectNameToFieldType = func() map[string]string { //nolint:gochecknoglobals
	m := make(map[string]string)
	for ft, obj := range fieldTypeToObjectName {
		m[obj] = ft
	}

	return m
}()

// objectsWithCustomFields is the set of objects that support custom fields.
var objectsWithCustomFields = datautils.NewStringSet( //nolint:gochecknoglobals
	"candidates", "applications", "jobs", "openings", "offers", "users",
)

// requestCustomFields fetches custom field definitions from the Greenhouse API for a given object.
// Returns a map of name_key to customFieldDefinition.
// For objects that do not support custom fields, an empty map is returned.
// https://harvestdocs.greenhouse.io/reference/get_v3-custom-fields
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]customFieldDefinition, error) {
	if !objectsWithCustomFields.Has(objectName) {
		return map[string]customFieldDefinition{}, nil
	}

	fieldType, ok := objectNameToFieldType[objectName]
	if !ok {
		return map[string]customFieldDefinition{}, nil
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", "custom_fields")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	url.WithQueryParam("field_type", fieldType)
	url.WithQueryParam("per_page", "500")

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[[]customFieldDefinition](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if fieldsResponse == nil {
		return map[string]customFieldDefinition{}, nil
	}

	fields := make(map[string]customFieldDefinition)

	for _, field := range *fieldsResponse {
		if !field.Active {
			continue
		}

		fields[field.NameKey] = field
	}

	return fields, nil
}

// customFieldDefinition represents a custom field definition from the Greenhouse API.
// https://harvestdocs.greenhouse.io/reference/get_v3-custom-fields
type customFieldDefinition struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	NameKey   string `json:"name_key"`
	FieldType string `json:"field_type"`
	ValueType string `json:"value_type"`
	Active    bool   `json:"active"`
	Private   bool   `json:"private"`
	Required  bool   `json:"required"`
}

// getValueType maps Greenhouse custom field value_type to common.ValueType.
// https://harvestdocs.greenhouse.io/reference/post_v3-custom-fields
func (f customFieldDefinition) getValueType() common.ValueType {
	switch f.ValueType {
	case "short_text", "long_text", "rich_text", "url":
		return common.ValueTypeString
	case "yes_no":
		return common.ValueTypeBoolean
	case "number":
		return common.ValueTypeInt
	case "date":
		return common.ValueTypeDateTime
	case "single_select":
		return common.ValueTypeSingleSelect
	default:
		// currency, currency_range, number_range, multi_select, user, linked, header, statement, attachment
		return common.ValueTypeOther
	}
}

// flattenCustomFields moves custom fields from the nested custom_fields map to the root level.
// In Greenhouse v3, custom_fields is a map keyed by name_key:
//
//	{
//	  "custom_fields": {
//	    "work_authorization": { "name": "Work Authorization", "type": "boolean", "value": true },
//	    "bio": { "name": "Bio", "type": "long_text", "value": "Some text" }
//	  }
//	}
//
// After flattening:
//
//	{
//	  "work_authorization": true,
//	  "bio": "Some text",
//	  ...original fields...
//	}
func flattenCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsRaw, ok := root["custom_fields"]
	if !ok || customFieldsRaw == nil {
		return root, nil
	}

	customFieldsMap, ok := customFieldsRaw.(map[string]any)
	if !ok {
		return root, nil
	}

	// Move each custom field value to the root level using name_key as the field name.
	for nameKey, fieldData := range customFieldsMap {
		fieldMap, ok := fieldData.(map[string]any)
		if !ok {
			continue
		}

		root[nameKey] = fieldMap["value"]
	}

	return root, nil
}

// requestCustomFieldOptions fetches options for single_select and multi_select custom fields.
// https://harvestdocs.greenhouse.io/reference/get_v3-custom-field-options
func (c *Connector) requestCustomFieldOptions(
	ctx context.Context, fieldIDs []int,
) (map[int][]customFieldOption, error) {
	if len(fieldIDs) == 0 {
		return map[int][]customFieldOption{}, nil
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "v3", "custom_field_options")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	// Build comma-separated list of field IDs.
	var idsSb strings.Builder

	for i, id := range fieldIDs {
		if i > 0 {
			idsSb.WriteString(",")
		}

		idsSb.WriteString(strconv.Itoa(id))
	}

	url.WithQueryParam("custom_field_ids", idsSb.String())
	url.WithQueryParam("per_page", "500")

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	optionsResponse, err := common.UnmarshalJSON[[]customFieldOption](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if optionsResponse == nil {
		return map[int][]customFieldOption{}, nil
	}

	// Group options by custom_field_id.
	result := make(map[int][]customFieldOption)

	for _, option := range *optionsResponse {
		if !option.Active {
			continue
		}

		result[option.CustomFieldID] = append(result[option.CustomFieldID], option)
	}

	return result, nil
}

// customFieldOption represents a select option for a custom field.
type customFieldOption struct {
	ID            int    `json:"id"`
	CustomFieldID int    `json:"custom_field_id"`
	Name          string `json:"name"`
	Active        bool   `json:"active"`
	SortOrder     int    `json:"sort_order"`
}
