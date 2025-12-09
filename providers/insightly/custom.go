// nolint:tagliatelle,godoclint
package insightly

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// Any object returned from READ operation may have a CUSTOMFIELDS array property.
// This struct is it's schema representation.
type readCustomField struct {
	Name  string `json:"FIELD_NAME"`
	Value any    `json:"FIELD_VALUE"`
}

/*
Provider object response:

	  {
		"RECORD_ID": 54840682,
		"RECORD_NAME": "Banana",
		"CUSTOMFIELDS": [
		  {
			"FIELD_NAME": "Color__c",
			"FIELD_VALUE": null,
			"CUSTOM_FIELD_ID": "Color__c"
		  },
		  {
			"FIELD_NAME": "Weight__c",
			"FIELD_VALUE": 3.2,
			"CUSTOM_FIELD_ID": "Weight__c"
		  }
		]
	  }

Read fields:

	  {
		"RECORD_ID": 54840682,
		"RECORD_NAME": "Banana",
		"Color__c": null,
		"Weight__c": 3.2
		"CUSTOMFIELDS": [...]
	  }
*/
func flattenCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsArray, err := jsonquery.New(node).ArrayOptional("CUSTOMFIELDS")
	if err != nil {
		return nil, err
	}

	if len(customFieldsArray) == 0 {
		// Nothing to move.
		return root, nil
	}

	// Custom fields are converted from the object representing the property to
	// key:value pair and moved to the top node level.
	for _, customFieldObject := range customFieldsArray {
		field, err := jsonquery.ParseNode[readCustomField](customFieldObject)
		if err != nil {
			return nil, err
		}

		root[field.Name] = field.Value
	}

	// Root level has adopted fields from custom fields.
	return root, nil
}

const customSuffix = "__c"

type customFieldsRegistry map[string]customFieldResponse

// Response from https://api.insightly.com/v3.1/Help#!/CustomFields/GetCustomFields.
type customFieldsResponse []customFieldResponse

type customFieldResponse struct {
	// FieldFor indicates the name of the object this custom field is associated with.
	// Use this to filter custom fields applicable to a specific object type.
	FieldFor    string `json:"FIELD_FOR"`
	Name        string `json:"FIELD_NAME"`
	DisplayName string `json:"FIELD_LABEL"`
	Type        string `json:"FIELD_TYPE"`
	Editable    bool   `json:"EDITABLE"`
	// Options holds the available choices for select-type field.
	Options []customFieldOption `json:"CUSTOM_FIELD_OPTIONS"`
}

// customFieldOption represents one of the enum option a custom field could take on.
// This applies for select-type fields.
type customFieldOption struct {
	// ID is an incremental number assigned in the order of creation in the dashboard.
	ID int `json:"OPTION_ID"`
	// Value is the label for one of the available choices in a select-type custom field.
	// For example, a field "Interests" may have values like "food", "sports", or "technology".
	Value string `json:"OPTION_VALUE"`
}

func (c customFieldResponse) BelongsToObject(objectName string) bool {
	first, _ := strings.CutSuffix(strings.ToLower(c.FieldFor), customSuffix)
	second, _ := strings.CutSuffix(strings.ToLower(objectName), customSuffix)

	return naming.NewSingularString(first).String() == // nolint:staticcheck
		naming.NewSingularString(second).String()
}

// requestCustomFields fetches custom fields for a given object from the Insightly API.
//
// Provider API Behavior:
// - If objectName supports custom fields and has them, the response is scoped to that object.
// - If objectName supports custom fields but doesn't have any, the API returns an empty list.
// - If objectName doesn't support custom fields or is unknown, API returns all system-wide custom fields.
// - If objectName is empty, API returns a 4xx error.
// - API treats singular and plural forms of objectName as equivalent.
// See: https://api.insightly.com/v3.1/Help#!/CustomFields/GetCustomFields
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (customFieldsRegistry, error) {
	if len(objectName) == 0 {
		return customFieldsRegistry{}, common.ErrEmptyObject
	}

	url, err := c.getCustomFieldsURL(objectName)
	if err != nil {
		return nil, err
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	customFields, err := common.UnmarshalJSON[customFieldsResponse](res)
	if err != nil {
		return nil, err
	}

	registry := make(customFieldsRegistry)

	for _, field := range *customFields {
		// Determine if the field is part of this object then add it.
		// API could return various fields so we must do a check.
		if field.BelongsToObject(objectName) {
			registry[field.Name] = field
		}
	}

	return registry, nil
}

// Response from https://api.insightly.com/v3.1/Help#!/CustomObjects/GetCustomObjects.
type customObjectResponse struct {
	ObjectName  string `json:"OBJECT_NAME"`
	DisplayName string `json:"PLURAL_LABEL"`
}

func (c *Connector) fetchCustomObject(
	ctx context.Context, objectName string,
) (*customObjectResponse, error) {
	url, err := c.getCustomObjectURL(objectName)
	if err != nil {
		return nil, err
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	customObject, err := common.UnmarshalJSON[customObjectResponse](res)
	if err != nil {
		return nil, err
	}

	return customObject, nil
}

// Response format from https://api.insightly.com/v3.1/Help#!/CustomObjectsRecords/GetEntities.
// This is a schema for custom objects. Every field that carries the data of interest
// will be part of `CUSTOMFIELDS` array. The fields for a custom object are defined in the dashboard.
var customObjectSchema = map[string]common.FieldMetadata{ // nolint:gochecknoglobals
	"RECORD_ID": {
		DisplayName:  "RECORD_ID",
		ValueType:    common.ValueTypeInt,
		ProviderType: "integer",
	},
	"RECORD_NAME": {
		DisplayName:  "RECORD_NAME",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
	},
	"DATE_CREATED_UTC": {
		DisplayName:  "DATE_CREATED_UTC",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
	},
	"DATE_UPDATED_UTC": {
		DisplayName:  "DATE_UPDATED_UTC",
		ValueType:    common.ValueTypeString,
		ProviderType: "string",
	},
	"CREATED_USER_ID": {
		DisplayName:  "CREATED_USER_ID",
		ValueType:    common.ValueTypeInt,
		ProviderType: "integer",
	},
	"VISIBLE_TO": {
		DisplayName:  "VISIBLE_TO",
		ValueType:    common.ValueTypeOther,
		ProviderType: "string",
	},
	"VISIBLE_TEAM_ID": {
		DisplayName:  "VISIBLE_TEAM_ID",
		ValueType:    common.ValueTypeOther,
		ProviderType: "string",
	},
	"CUSTOMFIELDS": {
		DisplayName:  "CUSTOMFIELDS",
		ValueType:    common.ValueTypeOther,
		ProviderType: "array",
	},
}
