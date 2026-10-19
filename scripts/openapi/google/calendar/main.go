package main

import (
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/openapi/google/internal/files"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi3"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		"*/watch",
		"/colors",
	}
	allowEndpoints = []string{
		"/calendars/{calendarId}/events",
		"/calendars/{calendarId}/acl",
	}
	displayNameOverride = map[string]string{
		"calendarList": "Calendars",
	}
	objectEndpoints = map[string]string{
		"/users/me/calendarList": "calendarList",
		"/users/me/settings":     "settings",
	}
	// nestedFieldParents lists object fields whose direct child properties are also exposed
	// as dotted field names (e.g. "start.dateTime") next to the parent object itself.
	//
	// The metadata model is flat, so this is the only way to let users map a single
	// nested value. Google requires start/end as objects and rejects a flat timestamp
	// mapped to "start"; the connector nests dotted keys on write and resolves them on read.
	// Keyed by object name, then the OpenAPI component that describes one record of it.
	nestedFieldParents = map[string]nestedFieldConfig{
		"events": {
			component: "Event",
			parents:   []string{"start", "end", "source", "extendedProperties"},
		},
	}
)

type nestedFieldConfig struct {
	component string
	parents   []string
}

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	document, err := files.CalendarDocument()
	goutils.MustBeNil(err)

	for _, object := range Objects() {
		urlPath := object.URLPath
		objectName := object.ObjectName

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", objectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			fieldMetadataMap := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  fieldNameConvertToDisplayName(field.Name),
					ValueType:    providerTypeConvertToValueType(field.Type),
					ProviderType: field.Type,
					Values:       nil,
				},
			}

			schemas.Add(providers.ModuleGoogleCalendar, objectName, object.DisplayName, urlPath,
				object.ResponseKey, fieldMetadataMap, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, objectName)
		}

		for fieldName, fieldMetadata := range nestedFields(document, objectName) {
			schemas.Add(providers.ModuleGoogleCalendar, objectName, object.DisplayName, urlPath,
				object.ResponseKey, staticschema.FieldMetadataMapV2{fieldName: fieldMetadata}, nil, object.Custom)
		}
	}

	goutils.MustBeNil(files.OutputCalendar.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputCalendar.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputCalendar.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects("GET",
		api3.AndPathMatcher{
			api3.NewDenyPathStrategy(ignoreEndpoints),
			api3.OrPathMatcher{
				api3.NewAllowPathStrategy(allowEndpoints),
				api3.IDPathIgnorer{},
			},
		},
		objectEndpoints, displayNameOverride,
		func(objectName, fieldName string) bool {
			return fieldName == "items"
		},
	)
	goutils.MustBeNil(err)

	return objects
}

func fieldNameConvertToDisplayName(fieldName string) string {
	return api3.CapitalizeFirstLetterEveryWord(
		api3.CamelCaseToSpaceSeparated(fieldName),
	)
}

func providerTypeConvertToValueType(providerType string) common.ValueType {
	switch providerType {
	case "integer":
		return common.ValueTypeInt
	case "string":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	default:
		// Ex: object, array
		return common.ValueTypeOther
	}
}

// nestedFields produces dotted field metadata for the configured parents of an object.
// Each direct child property of a parent becomes "<parent>.<child>" with the display name
// "<Parent> <Child>". Children that are themselves objects (e.g. extendedProperties.private,
// a free-form map) are emitted as objects without expanding further.
func nestedFields(document *openapi3.T, objectName string) staticschema.FieldMetadataMapV2 {
	config, ok := nestedFieldParents[objectName]
	if !ok {
		return nil
	}

	component, ok := document.Components.Schemas[config.component]
	if !ok || component.Value == nil {
		slog.Error("nested fields: component not found", "objectName", objectName, "component", config.component)

		return nil
	}

	result := staticschema.FieldMetadataMapV2{}

	for _, parent := range config.parents {
		parentSchema, ok := component.Value.Properties[parent]
		if !ok || parentSchema.Value == nil {
			slog.Error("nested fields: parent property not found", "objectName", objectName, "parent", parent)

			continue
		}

		for child, childSchema := range parentSchema.Value.Properties {
			if childSchema.Value == nil {
				continue
			}

			providerType := schemaProviderType(childSchema.Value)
			result[parent+"."+child] = staticschema.FieldMetadata{
				DisplayName: fmt.Sprintf("%v %v",
					fieldNameConvertToDisplayName(parent), fieldNameConvertToDisplayName(child)),
				ValueType:    leafValueType(providerType, childSchema.Value.Format),
				ProviderType: providerType,
				Values:       nil,
			}
		}
	}

	return result
}

func schemaProviderType(schema *openapi3.Schema) string {
	if schema.Type == nil || len(schema.Type.Slice()) == 0 {
		return "object"
	}

	return schema.Type.Slice()[0]
}

// leafValueType refines the value type of a nested leaf using the OpenAPI format hint,
// so that "start.dateTime" is a datetime and "start.date" is a date rather than plain strings.
func leafValueType(providerType, format string) common.ValueType {
	switch format {
	case "date":
		return common.ValueTypeDate
	case "date-time":
		return common.ValueTypeDateTime
	default:
		return providerTypeConvertToValueType(providerType)
	}
}
