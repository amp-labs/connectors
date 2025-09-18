package zendesksupport

import (
	"context"
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// requestCustomTicketFields makes and API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// The mapping is between "custom field id" and struct containing "human-readable field name".
//
// Custom fields are always associated with "ticket_fields" regardless of the object type.
func (c *Connector) requestCustomTicketFields(
	ctx context.Context, objectName string,
) (map[int64]ticketField, error) {
	if !objectsWithCustomFields[common.ModuleRoot].Has(objectName) {
		// This object doesn't have custom fields, we are done.
		return map[int64]ticketField{}, nil
	}

	return c.fetchCustomTicketFields(ctx)
}

func (c *Connector) fetchCustomTicketFields(ctx context.Context) (map[int64]ticketField, error) {
	url, err := c.getReadURL("ticket_fields")
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[ticketFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fields := make(map[int64]ticketField)
	for _, field := range fieldsResponse.TicketFields {
		fields[field.ID] = field
	}

	return fields, nil
}

// nolint:tagliatelle
// https://developer.zendesk.com/api-reference/ticketing/tickets/ticket_fields/#json-format
type ticketFieldsResponse struct {
	TicketFields []ticketField `json:"ticket_fields"`
}

// nolint:tagliatelle
type ticketField struct {
	// Automatically assigned when created
	ID int64 `json:"id"`
	// System or custom field type.
	Type string `json:"type"`
	// The title of the ticket field for end users in Help Center
	TitleInPortal string `json:"title_in_portal"`
	// Title is usually similar to the TitleInPortal.
	Title string `json:"title"`

	// Presented for a system ticket field of type "tickettype", "priority" or "status".
	SystemFieldOptions []systemFieldOption `json:"system_field_options,omitempty"`
	// List of customized ticket statuses. Only presented for a system ticket field of type "custom_status".
	CustomStatuses []customStatus `json:"custom_statuses,omitempty"`
	// Required and presented for a custom ticket field of type "multiselect" or "tagger".
	CustomFieldOptions []customFieldOption `json:"custom_field_options,omitempty"`
}

func (f ticketField) GetValueType() common.ValueType {
	switch f.Type {
	case "subject", "description":
		return common.ValueTypeString
	case
		// custom_field_options:
		"tagger",
		// system_field_options:
		"tickettype", "priority", "status",
		// custom_statuses:
		"custom_status":
		return common.ValueTypeSingleSelect
	case
		// custom_field_options:
		"multiselect":
		return common.ValueTypeMultiSelect
	default:
		// group, assignee
		return common.ValueTypeOther
	}
}

func (f ticketField) getValues() []common.FieldValue {
	result := make([]common.FieldValue, 0)

	for _, option := range f.SystemFieldOptions {
		result = append(result, common.FieldValue{
			Value:        option.Value,
			DisplayValue: option.Name,
		})
	}

	for _, option := range f.CustomFieldOptions {
		result = append(result, common.FieldValue{
			Value:        option.Value,
			DisplayValue: option.Name,
		})
	}

	for _, status := range f.CustomStatuses {
		result = append(result, common.FieldValue{
			Value:        strconv.FormatInt(status.ID, 10),
			DisplayValue: status.AgentLabel,
		})
	}

	return result
}

// nolint:tagliatelle
// https://developer.zendesk.com/api-reference/ticketing/tickets/custom_ticket_statuses/#json-format
type customStatus struct {
	ID int64 `json:"id"`
	// The label displayed to agents. Maximum length is 48 characters
	AgentLabel string `json:"agent_label"`
}

// https://developer.zendesk.com/api-reference/ticketing/tickets/ticket_fields/#list-ticket-field-options
type customFieldOption struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// https://developer.zendesk.com/api-reference/ticketing/tickets/ticket_fields/#list-ticket-field-options
type systemFieldOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// nolint:tagliatelle,lll
// https://developer.zendesk.com/documentation/ticketing/managing-tickets/creating-and-updating-tickets/#setting-custom-field-values
type readCustomFieldsResponse struct {
	CustomFields []readCustomField `json:"custom_fields"`
}

type readCustomField struct {
	ID    int64 `json:"id"`
	Value any   `json:"value"`
}

// Before parsing the records, if any custom fields are present (without a human-readable name),
// this will call the correct API to extend & replace the custom field with human-readable information.
// Object will then be enhanced using model.
func (c *Connector) attachReadCustomFields(
	customFields map[int64]ticketField,
) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if len(customFields) == 0 {
			// No custom fields, no-op, return as is.
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return enhanceObjectsWithCustomFieldNames(node, customFields)
	}
}

// In general this does the usual JSON parsing.
// However, those objects that contain "custom_fields" are processed as follows:
// * Locate custom fields in JSON read response.
// * Replace ids with human-readable names, which is provided as argument.
// * Place fields at the top level of the object.
func enhanceObjectsWithCustomFieldNames(
	node *ajson.Node,
	fields map[int64]ticketField,
) (map[string]any, error) {
	object, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsResponse, err := jsonquery.ParseNode[readCustomFieldsResponse](node)
	if err != nil {
		return nil, err
	}

	// Replace identifiers with human-readable field names which were found by making a call to "/model".
	for _, field := range customFieldsResponse.CustomFields {
		if model, ok := fields[field.ID]; ok {
			object[model.Title] = field.Value
		}
	}

	return object, nil
}
