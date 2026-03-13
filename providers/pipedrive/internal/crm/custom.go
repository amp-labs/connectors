package crm

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// objectsWithCustomFields defines which objects support custom fields in Pipedrive v2 API.
// These objects have a nested "custom_fields" object in read responses and
// corresponding *Fields endpoints for field definitions.
var objectsWithCustomFields = datautils.NewStringSet( //nolint:gochecknoglobals
	"activities",
	"deals",
	"persons",
	"organizations",
	"products",
)

// customFieldDef holds the definition of a single custom field.
type customFieldDef struct {
	Code      string // hash key, e.g. "a1b2c3d4..."
	Name      string // human-readable display name
	FieldType string // provider type, e.g. "varchar", "enum", "date"
}

// ValueType maps the Pipedrive field type to common.ValueType.
func (f customFieldDef) ValueType() common.ValueType {
	return nativeValueType(f.FieldType)
}

// requestCustomFields fetches the field definitions for an object from the
// matching *Fields endpoint and returns a map of hash→customFieldDef for
// custom fields only.
//
// Returns an empty map if the object doesn't support custom fields.
// Returns an error wrapped with common.ErrResolvingCustomFields on failure.
func (a *Adapter) requestCustomFields(
	ctx context.Context, objectName string,
) (map[string]customFieldDef, error) {
	if !objectsWithCustomFields.Has(objectName) {
		// This object doesn't support custom fields.
		return make(map[string]customFieldDef), nil
	}

	endpoint, ok := metadataDiscoveryEndpoints[objectName]
	if !ok {
		// No field definitions endpoint for this object.
		return make(map[string]customFieldDef), nil
	}

	url, err := a.getAPIURL(endpoint)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return nil, errors.Join(common.ErrResolvingCustomFields, err)
	}

	defs := make(map[string]customFieldDef)

	for _, fld := range response.Data {
		if fld.IsCustom {
			defs[fld.Code] = customFieldDef{
				Code:      fld.Code,
				Name:      fld.Name,
				FieldType: fld.FieldType,
			}
		}
	}

	return defs, nil
}

// attachReadCustomFields returns a RecordTransformer that promotes values from
// the nested "custom_fields" object to the root level using human-readable
// display names as keys.
//
// For objects that don't support custom fields, it returns the record as-is.
func (a *Adapter) attachReadCustomFields(
	objectName string, defs map[string]customFieldDef,
) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		if !objectsWithCustomFields.Has(objectName) {
			// This object doesn't support custom fields, return as-is.
			return jsonquery.Convertor.ObjectToMap(node)
		}

		return flattenCustomFields(node, defs)
	}
}

// flattenCustomFields promotes values from the nested "custom_fields" object
// to the root level, replacing hash keys with human-readable display names.
func flattenCustomFields(node *ajson.Node, defs map[string]customFieldDef) (map[string]any, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	customFieldsValue, ok := root["custom_fields"]
	if !ok {
		// No custom_fields in response.
		return root, nil
	}

	customFieldsMap, ok := customFieldsValue.(map[string]any)
	if !ok || len(customFieldsMap) == 0 {
		// custom_fields is not a map or is empty.
		return root, nil
	}

	// Promote each custom field value to root level with display name as key.
	for hash, value := range customFieldsMap {
		if def, found := defs[hash]; found {
			root[def.Name] = value
		} else {
			// Keep the hash key if no definition was found.
			root[hash] = value
		}
	}

	return root, nil
}
