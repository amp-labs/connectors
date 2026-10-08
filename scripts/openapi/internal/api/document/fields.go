package document

import (
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/getkin/kin-openapi/openapi3"
)

// fieldDefinitionLocation represents a position in an OpenAPI schema tree.
//
// Each location corresponds to a single schema object in the OpenAPI file.
// The paths slice holds the sequence of $ref URIs that must be followed from
// the root schema to reach this location, e.g.:
//
//	[]string{"#/components/schemas/User", "#/components/schemas/Address"}
type fieldDefinitionLocation struct {
	// schema is the OpenAPI schema object at this location.
	schema *openapi3.Schema

	// paths is the ordered list of $ref URIs from the root schema to this
	// schema. Each element is a full $ref string such as
	// "#/components/schemas/CompanyReference".
	paths []string
}

// newRootFieldLocation creates the root fieldDefinitionLocation for recursion.
//
// rootSchema is the top-level schema object from which field extraction starts.
// path is the initial $ref URI that identifies this root schema within the
// OpenAPI document (for example, "#/components/schemas/User").
func newRootFieldLocation(rootSchema *openapi3.Schema, path string) fieldDefinitionLocation {
	return fieldDefinitionLocation{
		schema: rootSchema,
		paths:  []string{path},
	}
}

// newNestedLocation returns a child location with an extra path element and
// a schema representing that child node in the OpenAPI schema tree.
func (l fieldDefinitionLocation) newNestedLocation(
	pathKey string,
	schema *openapi3.Schema,
) fieldDefinitionLocation {
	pathsCopy := slices.Clone(l.paths)
	pathsCopy = append(pathsCopy, pathKey)

	return fieldDefinitionLocation{
		schema: schema,
		paths:  pathsCopy,
	}
}

// extractFields walks the OpenAPI schema tree starting from location and
// collects all discoverable fields into a spec.Fields map.
//
// endpointPath is the API endpoint path associated with these fields (e.g.
// "/users" or "/companies/{id}"). propertyFlattener reports whether a given
// property should be treated as a container whose inner fields are lifted up
// rather than exposed as a nested object.
//
// origin is the original root schema from which this traversal began; it is
// stored in each FieldOrigin so that callers can always trace a field back to
// its top-level schema.
//
// location represents the current node in the schema tree. Its paths field
// encodes how to reach this node from the root by following $ref URIs. As
// extractFields recurses through anyOf, allOf, and properties, it creates
// child locations via newNestedLocation, extending this reference path.
//
// The function handles:
//   - anyOf: merges fields from all alternative schemas, representing the
//     union of possible fields.
//   - allOf: collects fields from all parent schemas, representing inherited
//     fields.
//   - properties: for each property, either:
//   - flattens its inner fields if propertyFlattener returns true, or
//   - records a top-level field with its type, enum options, and origin.
//
// The resulting spec.Fields map is keyed by property name and contains enough
// metadata (including FieldOrigin with ItemsSchema, Schema, RefPath, and
// ComponentRefs) to reconstruct how each field is defined and referenced in
// the OpenAPI document.
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
			nestedLocation := location.newNestedLocation(ref.Ref, ref.Value)
			fields := extractFields(endpointPath, propertyFlattener, origin, nestedLocation)
			combinedFields.AddMapValues(fields)
		}
	}

	// for all parents that exist collect fields
	for _, ref := range location.schema.AllOf {
		parentValue := ref.Value
		if parentValue != nil {
			nestedLocation := location.newNestedLocation(ref.Ref, parentValue)
			fields := extractFields(endpointPath, propertyFlattener, origin, nestedLocation)
			combinedFields.AddMapValues(fields)
		}
	}

	// properties local to this schema
	for property, propertySchema := range location.schema.Properties {
		if propertyFlattener(endpointPath, property) {
			// This property holds an array, and we need nested fields to be moved one level up.
			nestedLocation := location.newNestedLocation(propertySchema.Ref, propertySchema.Value)
			fields := extractFields(endpointPath, propertyFlattener, origin, nestedLocation)
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
