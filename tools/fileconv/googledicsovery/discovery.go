package googledicsovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/metadatadef"
)

type Document struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	XResources  map[string]Resource `json:"x-resources"` // nolint:tagliatelle
	XSchemas    map[string]Schema   `json:"x-schemas"`   // nolint:tagliatelle
}

func (d Document) ListObjects(httpMethod string) metadatadef.Schemas {
	objects := make(metadatadef.Schemas, 0)

	for objectName, resource := range d.XResources {
		for methodName, method := range resource.Methods {
			if methodName != "list" {
				continue
			}

			if !strings.EqualFold(method.HTTPMethod, httpMethod) {
				continue
			}

			if strings.Contains(method.Path, "{") {
				// No path IDs are allowed.
				continue
			}

			fields, err := d.schemaFieldsFor(method)

			objects = append(objects, metadatadef.Schema{
				ObjectName:  objectName,
				DisplayName: objectName,
				Fields:      fields,
				QueryParams: nil,
				URLPath:     method.Path,
				ResponseKey: "items",
				Problem:     err,
			})
		}
	}

	return objects
}

var (
	ErrMissingItemsRef = errors.New("missing nested reference to items schema")
	ErrUnknownRef      = errors.New("unknown schema reference")
)

func (d Document) schemaFieldsFor(method Method) (metadatadef.Fields, error) {
	schema, err := d.locateItemsSchema(method)
	if err != nil {
		return nil, err
	}

	return schema.fields(), nil
}

func (d Document) locateItemsSchema(method Method) (*Schema, error) {
	schema, err := d.findSchema(method.Response.Ref)
	if err != nil {
		return nil, err
	}

	itemsRef, ok := schema.Properties["items"]
	if !ok {
		return nil, fmt.Errorf("%w: problematic method is %v", ErrMissingItemsRef, method.Path)
	}

	bytes, err := json.Marshal(itemsRef)
	if err != nil {
		return nil, err
	}

	var items ItemsProperty
	if err = json.Unmarshal(bytes, &items); err != nil {
		return nil, err
	}

	return d.findSchema(items.Items.Ref)
}

func (d Document) findSchema(ref string) (*Schema, error) {
	schemaName, _ := strings.CutPrefix(ref, "$")
	for name, schema := range d.XSchemas {
		if name == schemaName {
			return &schema, nil
		}
	}

	return nil, fmt.Errorf("%w: reference %v", ErrUnknownRef, ref)
}

type Resource struct {
	Methods map[string]Method `json:"methods"`
}

type Method struct {
	HTTPMethod  string `json:"httpMethod"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Response    struct {
		Ref string `json:"$ref"`
	} `json:"response"`
}

type Schema struct {
	Id         string              `json:"id"`
	Properties map[string]Property `json:"properties"`
	Type       string              `json:"type"`
}

type Property struct {
	Type        string         `json:"type"`
	Items       map[string]any `json:"items"`
	Description string         `json:"description"`
	Ref         string         `json:"$ref"`
}

func (s *Schema) fields() metadatadef.Fields {
	fields := make(metadatadef.Fields)

	for propertyName, property := range s.Properties {
		propertyType := property.Type
		if len(propertyType) == 0 && len(property.Ref) != 0 {
			// Embedded object.
			propertyType = "object"
		}

		fields[propertyName] = metadatadef.Field{
			Name: propertyName,
			Type: propertyType,
		}
	}

	return fields
}

type ItemsProperty struct {
	Description string `json:"description"`
	Items       struct {
		Ref string `json:"$ref"`
	} `json:"items"`
	Type string `json:"type"`
}
