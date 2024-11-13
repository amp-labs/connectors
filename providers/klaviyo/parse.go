package klaviyo

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Every object has special field attributes which holds all the object specific fields.
// Therefore, nested "attributes" will be removed and fields will be moved to the top level of the object.
//
// Example accounts(shortened response):
//
//	 "data": [
//	    {
//	        "type": "",
//	        "id": "",
//	        "attributes": {
//	            "test_account": false,
//	            "contact_information": {},
//	            "locale": ""
//	        },
//	        "links": {}
//	    }
//	],
//
// The resulting fields for the above will be: type, id, test_account, contact_information, locale, links.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).Array("data", true)
	if err != nil {
		return nil, err
	}

	return flattenRecords(arr)
}

func flattenRecords(arr []*ajson.Node) ([]map[string]any, error) {
	result := make([]map[string]any, len(arr))

	for index, element := range arr {
		const keyAttributes = "attributes"

		attributes, err := jsonquery.New(element).Object(keyAttributes, true)
		if err != nil {
			return nil, err
		}

		original, err := jsonquery.Convertor.ObjectToMap(element)
		if err != nil {
			return nil, err
		}

		nested, err := jsonquery.Convertor.ObjectToMap(attributes)
		if err != nil {
			return nil, err
		}

		// Attributes object must be removed.
		delete(original, keyAttributes)

		// Fields from attributes are moved to the top level.
		for key, value := range nested {
			original[key] = value
		}

		result[index] = original
	}

	return result, nil
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}
