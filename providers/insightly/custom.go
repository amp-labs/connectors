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
// nolint:tagliatelle
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

const customMarker = "__c"

type customFieldsRegistry map[string]customFieldResponse

type customFieldsResponse []customFieldResponse

// nolint:tagliatelle
type customFieldResponse struct {
	Name        string `json:"FIELD_NAME"`
	DisplayName string `json:"FIELD_LABEL"`
	Type        string `json:"FIELD_TYPE"`
	Editable    bool   `json:"EDITABLE"`
	FieldFor    string `json:"FIELD_FOR"`
	Options     []struct {
		ID    int    `json:"OPTION_ID"`
		Value string `json:"OPTION_VALUE"`
	} `json:"CUSTOM_FIELD_OPTIONS"`
}

func (c customFieldResponse) BelongsToObject(objectName string) bool {
	if c.FieldFor == objectName {
		return true
	}

	if strings.ToLower(c.FieldFor) == strings.ToLower(objectName) { // nolint:staticcheck
		return true
	}

	first, _ := strings.CutSuffix(c.FieldFor, customMarker)
	second, _ := strings.CutSuffix(objectName, customMarker)

	return strings.ToLower(naming.NewSingularString(first).String()) == // nolint:staticcheck
		strings.ToLower(naming.NewSingularString(second).String())
}

// requestCustomFields fetches custom fields for a given object from the Insightly API.
//
// Behavior:
// - If objectName supports custom fields and has them, the response is scoped to that object.
// - If objectName supports custom fields but doesn't have any, the API returns an empty list.
// - If objectName doesn't support custom fields or is unknown, API returns all system-wide custom fields.
// - If objectName is empty, API returns a 4xx error.
// - API treats singular and plural forms of objectName as equivalent.
// See: https://api.insightly.com/v3.1/Help#!/CustomFields/GetCustomFields
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (customFieldsRegistry, error) {
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

// nolint:tagliatelle
type customObjectResponse struct {
	ObjectName  string `json:"OBJECT_NAME"`
	DisplayName string `json:"PLURAL_LABEL"`
}

func (c *Connector) fetchCustomObjectDisplayName(
	ctx context.Context, objectName string,
) (string, error) {
	url, err := c.getCustomObjectURL(objectName)
	if err != nil {
		return "", err
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	customObject, err := common.UnmarshalJSON[customObjectResponse](res)
	if err != nil {
		return "", err
	}

	return customObject.DisplayName, nil
}

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
