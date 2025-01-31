package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/google"
	"github.com/amp-labs/connectors/providers/google/discovery"
	"github.com/amp-labs/connectors/providers/google/metadata"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/fileconv/googledicsovery"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	displayNameOverride = map[string]string{
		"acl":          "ACL",
		"calendarList": "Calendars",
	}
)

func main() {
	explorer, err := discovery.CalendarFileManager.GetExplorer(
		googledicsovery.WithDisplayNamePostProcessors(
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(displayNameOverride)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			path := "calendar/v3/" + object.URLPath
			schemas.Add(google.ModuleCalendar, object.ObjectName, object.DisplayName, path, object.ResponseKey,
				staticschema.FieldMetadataMapV2{
					field.Name: staticschema.FieldMetadata{
						DisplayName:  fieldNameConvertToDisplayName(field.Name),
						ValueType:    providerTypeConvertToValueType(field.Type),
						ProviderType: field.Type,
						ReadOnly:     false,
						Values:       nil,
					},
				}, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
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
