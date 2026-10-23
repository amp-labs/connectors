package main

import (
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/jump/metadata"
	"github.com/amp-labs/connectors/providers/jump/metadata/graphql"
)

const (
	connectionSuffix = "Connection"
	itemsField       = "items"
)

var errMissingQueryRoot = errors.New("schema has no query root type")

// nolint:gochecknoglobals
var (
	scalarValueTypes = map[string]common.ValueType{
		"ID":       common.ValueTypeString,
		"String":   common.ValueTypeString,
		"Int":      common.ValueTypeInt,
		"Float":    common.ValueTypeFloat,
		"Boolean":  common.ValueTypeBoolean,
		"DateTime": common.ValueTypeDateTime,
		"Date":     common.ValueTypeDate,
	}

	blockStrings  = regexp.MustCompile(`(?s)""".*?"""`)
	lineStrings   = regexp.MustCompile(`"(?:[^"\\\n]|\\.)*"`)
	arguments     = regexp.MustCompile(`\([^()]*\)`)
	queryRootType = regexp.MustCompile(`schema\s*\{[^}]*query:\s*(\w+)`)
	definitions   = regexp.MustCompile(`\b(type|enum)\s+(\w+)[^{]*\{([^}]*)\}`)
)

// definition is a GraphQL object type or enum declared in the SDL.
type definition struct {
	isEnum     bool
	fields     []field // object type fields, in SDL order
	enumValues []string
}

type field struct {
	name string
	// typ is the type as written in the SDL, e.g. "ID!" or "[ContactInfo]".
	typ string
}

// Jump disables GraphQL introspection, so object metadata is generated from the published SDL.
// Every list query returning a "<Type>Connection" becomes an object, and the fields of its
// "items" type become the object fields.
func main() {
	queryRoot, types := parseSDL(string(graphql.SchemaFile))

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()

	for _, query := range types[queryRoot].fields {
		connectionName := namedType(query.typ)
		if !strings.HasSuffix(connectionName, connectionSuffix) {
			continue
		}

		items, ok := fieldByName(types[connectionName], itemsField)
		if !ok {
			slog.Error("connection has no items field", "query", query.name, "type", connectionName)

			continue
		}

		objectName := query.name
		displayName := naming.CapitalizeFirstLetterEveryWord(naming.SeparateCamelCaseWords(objectName))

		for _, itemField := range types[namedType(items.typ)].fields {
			schemas.Add("", objectName, displayName, "", itemsField,
				staticschema.FieldMetadataMapV2{
					itemField.name: convertField(types, itemField),
				}, nil, nil)
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))

	slog.Info("Completed.")
}

// parseSDL reads the type and enum declarations Jump's schema uses.
// Descriptions and field arguments are not needed for metadata, so they are removed first.
func parseSDL(source string) (string, map[string]*definition) {
	source = blockStrings.ReplaceAllString(source, "")
	source = lineStrings.ReplaceAllString(source, "")

	for arguments.MatchString(source) {
		source = arguments.ReplaceAllString(source, "")
	}

	types := make(map[string]*definition)

	for _, match := range definitions.FindAllStringSubmatch(source, -1) {
		kind, name, body := match[1], match[2], match[3]

		if kind == "enum" {
			types[name] = &definition{isEnum: true, enumValues: strings.Fields(body)}

			continue
		}

		types[name] = &definition{fields: parseFields(body)}
	}

	queryRoot := queryRootType.FindStringSubmatch(source)
	if queryRoot == nil {
		goutils.MustBeNil(errMissingQueryRoot)
	}

	return queryRoot[1], types
}

func parseFields(body string) []field {
	var fields []field

	for line := range strings.SplitSeq(body, "\n") {
		name, typ, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields = append(fields, field{
			name: strings.TrimSpace(name),
			typ:  strings.Fields(typ)[0],
		})
	}

	return fields
}

func fieldByName(object *definition, name string) (field, bool) {
	if object == nil {
		return field{}, false
	}

	for _, candidate := range object.fields {
		if candidate.name == name {
			return candidate, true
		}
	}

	return field{}, false
}

// namedType strips list and non-null markers: "[ContactInfo!]!" becomes "ContactInfo".
func namedType(typ string) string {
	return strings.Trim(typ, "[]!")
}

func isList(typ string) bool {
	return strings.HasPrefix(typ, "[")
}

func convertField(types map[string]*definition, sdlField field) staticschema.FieldMetadata {
	return staticschema.FieldMetadata{
		DisplayName:  sdlField.name,
		ValueType:    getFieldValueType(types, sdlField.typ),
		ProviderType: sdlField.typ,
		Values:       getFieldValueOptions(types, sdlField.typ),
	}
}

func getFieldValueType(types map[string]*definition, typ string) common.ValueType {
	enum := isEnum(types, typ)

	if isList(typ) {
		if enum {
			return common.ValueTypeMultiSelect
		}

		return common.ValueTypeOther
	}

	if enum {
		return common.ValueTypeSingleSelect
	}

	if valueType, ok := scalarValueTypes[namedType(typ)]; ok {
		return valueType
	}

	// Objects and the arbitrary Json scalar.
	return common.ValueTypeOther
}

func getFieldValueOptions(types map[string]*definition, typ string) staticschema.FieldValues {
	if !isEnum(types, typ) {
		return nil
	}

	enumValues := types[namedType(typ)].enumValues

	values := make(staticschema.FieldValues, len(enumValues))
	for index, enumValue := range enumValues {
		values[index] = staticschema.FieldValue{
			Value:        enumValue,
			DisplayValue: enumValue,
		}
	}

	return values
}

func isEnum(types map[string]*definition, typ string) bool {
	definition, ok := types[namedType(typ)]

	return ok && definition.isEnum
}
