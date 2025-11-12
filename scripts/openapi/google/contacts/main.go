package main

import (
	"log/slog"
	"net/http"
	"strings"

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

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// Requires query parameter `resourceNames=[]`.
		"/v1/contactGroups:batchGet",
		"/v1/people:batchGet",
		// Requires query parameter `query`, that is a text search.
		"/v1/otherContacts:search",
		"/v1/people:searchContacts",
		"/v1/people:searchDirectoryPeople",
		// URL that require IDs.
		// We are explicit, because we need some such endpoints.
		// The IDs will be hard coded making them regular endpoints.
		"/v1/{resourceName}",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/v1/people:listDirectoryPeople": "peopleDirectory",
	}
)

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range Objects() {
		urlPath, _ := strings.CutPrefix(object.URLPath, "/v1")
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

			schemas.Add(providers.ModuleGoogleContacts, objectName, object.DisplayName, urlPath,
				object.ResponseKey, fieldMetadataMap, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(files.OutputContacts.FlushSchemas(schemas))
	goutils.MustBeNil(files.OutputContacts.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func Objects() []metadatadef.Schema {
	explorer, err := files.InputContacts.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjects(
		http.MethodGet,
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, nil,
		func(objectName, fieldName string) bool {
			return false
		},
	)
	goutils.MustBeNil(err)

	for index, object := range objects {
		if object.URLPath == "/v1/{resourceName}/connections" {
			// Override some values.
			object.URLPath = "/v1/people/me/connections"
			object.ObjectName = "myConnections"
			objects[index] = object
		}
	}

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
