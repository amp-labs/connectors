package insightly

import (
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
