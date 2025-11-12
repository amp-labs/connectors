package main

import (
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
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

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
