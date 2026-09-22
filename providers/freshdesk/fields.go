package freshdesk

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// fieldDefinitionResource names the endpoint declaring an object's fields.
// Freshdesk describes only these three objects, the rest are known from their payload alone.
var fieldDefinitionResource = map[string]string{ //nolint:gochecknoglobals
	"companies": "company_fields",
	"contacts":  "contact_fields",
	"tickets":   "ticket_fields",
}

// declaredFieldTypes maps the types Freshdesk names onto Ampersand types.
// A type that is absent leaves the sampled type untouched, which keeps a declaration we cannot
// interpret from overruling what the data plainly shows. Lists such as domains or tag_names sit
// here deliberately: they have no Ampersand equivalent and the sampled type already says so.
var declaredFieldTypes = map[string]common.ValueType{ //nolint:gochecknoglobals
	"custom_text":                common.ValueTypeString,
	"custom_paragraph":           common.ValueTypeString,
	"custom_url":                 common.ValueTypeString,
	"custom_phone_number":        common.ValueTypeString,
	"default_address":            common.ValueTypeString,
	"default_description":        common.ValueTypeString,
	"default_email":              common.ValueTypeString,
	"default_job_title":          common.ValueTypeString,
	"default_mobile":             common.ValueTypeString,
	"default_name":               common.ValueTypeString,
	"default_note":               common.ValueTypeString,
	"default_phone":              common.ValueTypeString,
	"default_subject":            common.ValueTypeString,
	"default_unique_external_id": common.ValueTypeString,

	"custom_number":  common.ValueTypeInt,
	"custom_decimal": common.ValueTypeFloat,

	"custom_checkbox": common.ValueTypeBoolean,

	"custom_date":          common.ValueTypeDate,
	"default_renewal_date": common.ValueTypeDate,

	"custom_dropdown":      common.ValueTypeSingleSelect,
	"default_account_tier": common.ValueTypeSingleSelect,
	"default_health_score": common.ValueTypeSingleSelect,
	"default_industry":     common.ValueTypeSingleSelect,
	"default_language":     common.ValueTypeSingleSelect,
	"default_priority":     common.ValueTypeSingleSelect,
	"default_source":       common.ValueTypeSingleSelect,
	"default_status":       common.ValueTypeSingleSelect,
	"default_ticket_type":  common.ValueTypeSingleSelect,
	"default_time_zone":    common.ValueTypeSingleSelect,
}

// declaredFieldAliases maps a declared field onto the field that records carry.
// Freshdesk names several declarations after the concept rather than the field holding it, and
// says nothing about which field that is. Since a declaration is what describes an object here,
// an unmapped one would be reported under a name no record carries, hence this table.
var declaredFieldAliases = map[string]string{ //nolint:gochecknoglobals
	// Tickets.
	"agent":       "responder_id",
	"company":     "company_id",
	"group":       "group_id",
	"product":     "product_id",
	"requester":   "requester_id",
	"ticket_type": "type",
	// Contacts.
	"company_name": "company_id",
	"list_ids":     "lists",
	"tag_names":    "tags",
}

// fieldDefinition is a field as Freshdesk declares it.
type fieldDefinition struct {
	Id                int64           `json:"id"` //nolint:revive
	Name              string          `json:"name"`
	Label             string          `json:"label"`
	Type              string          `json:"type"`
	Default           bool            `json:"default"`
	RequiredForAgents bool            `json:"required_for_agents"`
	Choices           json.RawMessage `json:"choices"`
}

// overlayDeclaredFields applies what Freshdesk declares about an object to the given fields.
// A declaration carries the type, label and choices that the sampled values cannot show.
//
// Where Freshdesk declares an object, its declarations describe it and the sampled records cover
// what they leave out, id and the timestamps among them. A declaration therefore contributes its
// field whether or not a record carried it, under the name records use for it, and its type wins
// over the one a value implied. Custom fields are the exception, skipped because a record carries
// those inside custom_fields rather than at its top level.
func (conn *Connector) overlayDeclaredFields(
	ctx context.Context, object string, fields common.FieldsMetadata,
) error {
	resource, known := fieldDefinitionResource[object]
	if !known {
		return nil
	}

	url, err := urlbuilder.New(conn.BaseURL, restAPIPrefix, resource)
	if err != nil {
		return err
	}

	response, err := conn.Client.Get(ctx, url.String())
	if err != nil {
		return err
	}

	definitions, err := common.UnmarshalJSON[[]fieldDefinition](response)
	if err != nil {
		return common.ErrFailedToUnmarshalBody
	}

	for _, definition := range *definitions {
		if !definition.Default {
			continue
		}

		name := definition.payloadName()

		field, sampled := fields[name]
		if !sampled {
			// Declared, yet absent from the sampled page. Its type rests on the declaration
			// alone, so one worded in a way we cannot read leaves the field untyped.
			field = common.FieldMetadata{DisplayName: name, ValueType: common.ValueTypeOther}
		}

		fields[name] = definition.refine(field)
	}

	return nil
}

// payloadName is the field of a record that this declaration describes.
func (d fieldDefinition) payloadName() string {
	if alias, aliased := declaredFieldAliases[d.Name]; aliased {
		return alias
	}

	return d.Name
}

// refine applies what Freshdesk declares on top of what the sampled value implied.
func (d fieldDefinition) refine(field common.FieldMetadata) common.FieldMetadata {
	custom := !d.Default //nolint:staticcheck // always false today, kept for when custom fields are read
	required := d.RequiredForAgents
	fieldId := strconv.FormatInt(d.Id, 10)

	field.ProviderType = d.Type
	field.IsCustom = &custom
	field.IsRequired = &required
	field.FieldId = &fieldId

	if d.Label != "" {
		field.DisplayName = d.Label
	}

	valueType, mapped := declaredFieldTypes[d.Type]
	if !mapped {
		return field
	}

	field.ValueType = valueType
	if valueType == common.ValueTypeSingleSelect || valueType == common.ValueTypeMultiSelect {
		field.Values = d.values()
	}

	return field
}

// values reads the options of a selection field. Freshdesk writes them in three shapes:
// a list of labels, a map of label to stored identifier, and a map of stored identifier to labels.
func (d fieldDefinition) values() []common.FieldValue {
	if len(d.Choices) == 0 {
		return nil
	}

	var labels []string
	if err := json.Unmarshal(d.Choices, &labels); err == nil {
		values := make([]common.FieldValue, 0, len(labels))
		for _, label := range labels {
			values = append(values, common.FieldValue{Value: label, DisplayValue: label})
		}

		return values
	}

	var mapping map[string]any
	if err := json.Unmarshal(d.Choices, &mapping); err != nil {
		return nil
	}

	values := make([]common.FieldValue, 0, len(mapping))

	for key, option := range mapping {
		values = append(values, choiceValue(key, option))
	}

	sort.Slice(values, func(a, b int) bool {
		return values[a].DisplayValue < values[b].DisplayValue
	})

	return values
}

// choiceValue resolves which side of a choice mapping is stored on the record.
// A numeric option is the stored identifier and the key is its label, whereas a textual option
// is the label of a key that is itself stored, as language does with "en" against "English".
// Status is the outlier, listing the agent and customer facing labels of a stored identifier.
func choiceValue(key string, option any) common.FieldValue {
	switch value := option.(type) {
	case float64:
		return common.FieldValue{
			Value:        strconv.FormatFloat(value, 'f', -1, 64),
			DisplayValue: key,
		}
	case string:
		return common.FieldValue{Value: key, DisplayValue: value}
	case []any:
		if len(value) != 0 {
			if label, ok := value[0].(string); ok {
				return common.FieldValue{Value: key, DisplayValue: label}
			}
		}
	}

	return common.FieldValue{Value: key, DisplayValue: key}
}
