package document

import (
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

type fieldDefinitionLocation struct {
	schema *openapi3.Schema
	paths  []string
}

func extractFields(
	endpointPath string,
	propertyFlattener PropertyFlattener,
	origin *openapi3.Schema,
	location fieldDefinitionLocation,
) spec.Fields {
	combinedFields := make(spec.Fields)

	if location.schema.AnyOf != nil {
		// this object can be represented by various definitions
		// we merge those fields to represent the whole domain of possible fields
		// of course omitting duplicates.
		for _, ref := range location.schema.AnyOf {
			fields := extractFields(endpointPath, propertyFlattener, origin, fieldDefinitionLocation{
				schema: ref.Value,
				paths:  append(location.paths, ref.Ref),
			})
			combinedFields.AddMapValues(fields)
		}
	}

	// for all parents that exist collect fields
	for _, ref := range location.schema.AllOf {
		parentValue := ref.Value
		if parentValue != nil {
			fields := extractFields(endpointPath, propertyFlattener, origin, fieldDefinitionLocation{
				schema: parentValue,
				paths:  append(location.paths, ref.Ref),
			})
			combinedFields.AddMapValues(fields)
		}
	}

	// properties local to this schema
	for property, propertySchema := range location.schema.Properties {
		if propertyFlattener(endpointPath, property) {
			// This property holds an array, and we need nested fields to be moved one level up.
			fields := extractFields(endpointPath, propertyFlattener, origin, fieldDefinitionLocation{
				schema: propertySchema.Value,
				paths:  append(location.paths, propertySchema.Ref),
			})
			combinedFields.AddMapValues(fields)
		} else {
			// This is just a normal usual case where top level fields are collected as is.
			propertyType := extractPropertyType(propertySchema)
			enumOptions := extractEnumOptions(endpointPath, propertySchema)

			primaryRef := singlePropertyComponentRef(propertySchema)
			componentRefs := propertyComponentRefs(propertySchema)

			combinedFields[property] = spec.Field{
				URLPath:      endpointPath,
				Name:         property,
				DisplayName:  property,
				Type:         propertyType,
				ValueOptions: enumOptions,
				Origin:       buildFieldOrigin(origin, location, primaryRef, componentRefs),
			}
		}
	}

	return combinedFields
}

// buildFieldOrigin constructs FieldOrigin with consistent ComponentRefs/ComponentNames.
func buildFieldOrigin(origin *openapi3.Schema,
	location fieldDefinitionLocation,
	primaryRef string,
	componentRefs []string,
) spec.FieldOrigin {
	componentNames := make([]string, len(componentRefs))
	for i, ref := range componentRefs {
		componentNames[i] = schemaRefComponentName(ref)
	}

	refPath := location.paths
	if primaryRef != "" {
		refPath = append(refPath, primaryRef)
	}

	return spec.FieldOrigin{
		ItemsSchema:    origin,
		Schema:         location.schema, // schema directly holding the field
		RefPath:        refPath,
		ComponentRefs:  componentRefs,
		ComponentNames: componentNames,
	}
}

func extractPropertyType(propertySchema *openapi3.SchemaRef) string {
	if propertySchema.Value == nil || propertySchema.Value.Type == nil {
		return ""
	}

	types := *propertySchema.Value.Type
	if len(types) == 0 {
		return ""
	}

	return types[0]
}

func extractEnumOptions(endpointPath string, propertySchema *openapi3.SchemaRef) []string {
	enumOptions := make([]string, 0)

	if propertySchema.Value != nil && propertySchema.Value.Enum != nil {
		for _, value := range propertySchema.Value.Enum {
			if option, ok := mapEnumToString(endpointPath, value); ok {
				enumOptions = append(enumOptions, option)
			}
		}
	}

	return enumOptions
}

func mapEnumToString(endpointPath string, value any) (string, bool) {
	if value == nil {
		return "null", true
	}

	switch option := value.(type) {
	case string:
		return option, true
	case float64:
		return strconv.FormatFloat(option, 'f', -1, 64), true
	default:
		slog.Warn("Enum option is not a string", "endpoint", endpointPath)
	}

	return "", false
}

// propertyComponentRefs resolves the OpenAPI component name the property schema referenced.
// For object properties it is the property's own $ref, for array properties the $ref of its items,
// even when the array is declared through a named wrapper schema.
// Empty string when the schema is inlined.
func propertyComponentRefs(propertySchema *openapi3.SchemaRef) []string {
	if propertySchema == nil {
		return nil
	}

	if propertySchema.Value != nil && propertySchema.Value.Items != nil {
		return componentReferences(propertySchema.Value.Items)
	}

	return componentReferences(propertySchema)
}

// singlePropertyComponentRef returns a single $ref if this schema has exactly one
// component reference (direct or via allOf/oneOf/anyOf). If there are 0 or > 1
// distinct component refs, it returns "".
func singlePropertyComponentRef(schema *openapi3.SchemaRef) string {
	refs := componentReferences(schema)
	if len(refs) == 1 {
		return refs[0]
	}

	return ""
}

// componentReferences returns all distinct component $refs that define this schema.
// It inspects $ref on the schema itself and within allOf/oneOf/anyOf.
// The result may have 0, 1, or many elements.
func componentReferences(schema *openapi3.SchemaRef) []string {
	if schema == nil {
		return nil
	}

	refs := datautils.NewSet[string]()

	// Direct ref
	if schema.Ref != "" {
		refs.AddOne(schema.Ref)
	}

	if schema.Value != nil {
		for _, composite := range datautils.MergeSlices(
			schema.Value.AllOf,
			schema.Value.OneOf,
			schema.Value.AnyOf,
		) {
			if composite == nil {
				continue
			}

			if composite.Ref != "" {
				refs.AddOne(composite.Ref)
			}
		}
	}

	result := refs.List()
	sort.Strings(result)

	return result
}

// schemaRefComponentName converts a $ref URI to the component name.
// Ex: "#/components/schemas/CompanyReference" -> "CompanyReference".
func schemaRefComponentName(ref string) string {
	if ref == "" {
		return ""
	}

	return ref[strings.LastIndex(ref, "/")+1:]
}
