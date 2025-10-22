package capsule

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
// List of object names which support custom fields.
// Docs on Custom fields: https://developer.capsulecrm.com/v2/operations/Custom_Field#listFields
var objectsWithCustomFields = datautils.NewStringSet("parties", "opportunities", "kases", "projects")

// requestCustomFields makes and API call to get model describing custom fields.
// For not applicable objects the empty mapping is returned.
// map[string]modelCustomField is a map of field name to the field definition itself.
func (c *Connector) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]modelCustomField, error) {
	if !objectsWithCustomFields.Has(objectName) {
		// This object doesn't have custom fields, we are done.
		return map[string]modelCustomField{}, nil
	}

	url, err := c.getCustomFieldsURLFor(objectName)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	fieldsResponse, err := common.UnmarshalJSON[modelCustomFieldsResponse](res)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	if fieldsResponse == nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, common.ErrEmptyJSONHTTPResponse)
	}

	fields := make(map[string]modelCustomField)
	for _, field := range fieldsResponse.Definitions {
		fields[field.Name] = field
	}

	return fields, nil
}

// nolint:tagliatelle
type modelCustomFieldsResponse struct {
	Definitions []modelCustomField `json:"definitions"`
}

// nolint:tagliatelle
type modelCustomField struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`
	Description  any    `json:"description"`
	Important    bool   `json:"important"`
	DisplayOrder int    `json:"displayOrder"`
	Name         string `json:"name"`
	Tag          struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		DataTag bool   `json:"dataTag"`
	} `json:"tag"`
	CaptureRule any      `json:"captureRule"`
	Options     []string `json:"options"`
}

// https://developer.capsulecrm.com/v2/models/field_definition
func (f modelCustomField) getValueType() common.ValueType {
	switch f.Type {
	case "text":
		return common.ValueTypeString
	case "date":
		return common.ValueTypeDate
	case "list":
		return common.ValueTypeSingleSelect
	case "boolean":
		return common.ValueTypeBoolean
	case "number":
		return common.ValueTypeFloat
	case "link":
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func (f modelCustomField) getValues() common.FieldValues {
	return datautils.ForEach(f.Options, func(option string) common.FieldValue {
		return common.FieldValue{
			Value:        option,
			DisplayValue: option,
		}
	})
}

func flattenCustomFields(node *ajson.Node) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsArray, err := jsonquery.New(node).ArrayOptional("fields")
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

		root[field.Definition.Name] = field.Value
	}

	// Root level has adopted fields from custom fields.
	return root, nil
}

// Custom field schema as it appears when reading "Projects", "Opportunities", "Kases" (aka Projects).
//
// Example:
//
//	{
//	 "kases": [{
//	   "id": 5588202,
//	   "fields": [
//	     {
//	       "id": 9785121,
//	       "definition": {
//	         "id": 926886,
//	         "name": "Interests"
//	       },
//	       "value": "Skiing",
//	       "tagId": 168298
//	     }
//	   ]
//	 }]
//	}
type readCustomField struct {
	ID         int `json:"id"`
	Definition struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"definition"`
	Value string `json:"value"`
	TagID int    `json:"tagId"`
}
