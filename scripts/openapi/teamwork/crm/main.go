package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/scripts/openapi/teamwork/internal/files"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{
		// Singular objects.
		"/projects/api/v3/companies/time.json",
		"/projects/api/v3/me.json",
		"/projects/api/v3/people/utilization.json",
		"/projects/api/v3/projects/metrics/active.json",
		"/projects/api/v3/projects/metrics/invoice.json",
		"/projects/api/v3/projects/metrics/unbilled.json",
		"/projects/api/v3/reporting/precanned/companytime.json",
		"/projects/api/v3/reporting/precanned/usertaskcompletion.json",
		"/projects/api/v3/summary.json",
		"/projects/api/v3/tasks/metrics/complete.json",
		"/projects/api/v3/tasks/metrics/late.json",
		"/projects/api/v3/time/total.json",
		"/projects/api/v3/timesheets/totals.json",
		"/projects/api/v3/workload.json",
	}
	displayNameOverride = map[string]string{}
	objectEndpoints     = map[string]string{
		"/projects/api/v3/me/timers": "me/timers",
		"/projects/api/v3/timers":    "timers",
	}
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		name := formatObjectName(object.ObjectName)
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", name,
				"error", object.Problem,
				"url", object.URLPath,
			)
		}

		for _, field := range object.Fields {
			fieldMetadataMap := staticschema.FieldMetadataMapV2{
				field.Name: staticschema.FieldMetadata{
					DisplayName:  fieldNameConvertToDisplayName(field.Name),
					ValueType:    providerTypeConvertToValueType(field.Type),
					ProviderType: field.Type,
					ReadOnly:     false,
					Values:       nil,
				},
			}

			path, _ := strings.CutPrefix(object.URLPath, "/projects/api/v3")

			schemas.Add(common.ModuleRoot, name, object.DisplayName, path,
				object.ResponseKey, fieldMetadataMap, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, name)
		}
	}

	goutils.MustBeNil(files.OutputTeamwork.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputTeamwork.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputTeamwork.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			formatDisplayName,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects("GET",
		api3.AndPathMatcher{
			api3.NewDenyPathStrategy(ignoreEndpoints),
			// There should be no curly brackets no IDs, no nested resources.
			// Read objects are those that have constant string path.
			colonIDIgnorer{},
			api3.IDPathIgnorer{},
		},
		objectEndpoints, displayNameOverride,
		func(objectName, fieldName string) bool {
			return formatObjectName(objectName) == fieldName
		},
	)
	goutils.MustBeNil(err)

	return objects
}

func formatObjectName(objectName string) string {
	objectName, _ = strings.CutPrefix(objectName, "projects/api/v3/")
	objectName, _ = strings.CutSuffix(objectName, ".json")

	return objectName
}

func formatDisplayName(displayName string) string {
	displayName, _ = strings.CutSuffix(displayName, " Response")
	displayName, _ = strings.CutSuffix(displayName, " Response V 205")

	return displayName
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

type colonIDIgnorer struct{}

func (colonIDIgnorer) IsPathMatching(path string) bool {
	return !strings.Contains(path, ":id")
}
