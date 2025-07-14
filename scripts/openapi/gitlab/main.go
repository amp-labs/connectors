package main

import (
	_ "embed"
	"log"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api2"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	ignoreEndpoints = []string{}
	objectEndpoints = map[string]string{
		"/api/v4/user/keys":                       "user/keys",
		"/api/v4/keys":                            "keys",
		"/api/v4/groups/import/authorize":         "groups/import/authorize",
		"/api/v4/projects/import/authorize":       "projects/import/authorize",
		"/api/v4/user/runners":                    "user/runners",
		"/api/v4/runners":                         "runners",
		"/api/v4/events":                          "events",
		"/api/v4/integrations/slack/events":       "integrations/slack/events",
		"/api/v4/container_registry_event/events": "container_registry_event/events",
		"/api/v4/user/status":                     "user/status",
		"/api/v4/geo/status":                      "geo/status",
		"/api/v4/groups/import":                   "groups/import",
		"/api/v4/projects/import":                 "projects/import",
		"/api/v4/user/personal_access_tokens":     "user/personal_access_tokens",
		"/api/v4/personal_access_tokens":          "personal_access_tokens",
		"/api/v4/snippets/all":                    "snippets/all",
		"/api/v4/runners/all":                     "runners/all",
	}
	displayNameOverride       = map[string]string{}
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{},
		func(objectName string) (fieldName string) {
			return ""
		},
	)

	// Static file containing openapi spec.
	//
	//go:embed openapi_v2.yaml
	apiFile     []byte
	FileManager = api2.NewOpenapiFileManager[any](apiFile) // nolint:gochecknoglobals

	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas       []byte
	schemaManager = scrapper.NewMetadataFileManager[staticschema.FieldMetadataMapV1]( // nolint:gochecknoglobals
		schemas, fileconv.NewSiblingFileLocator())

	// Schemas is cached Object schemas.
	Schemas = schemaManager.MustLoadSchemas() // nolint:gochecknoglobals
)

func main() {
	explorer, err := FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		))
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV1]()
	registry := datautils.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		objectName, _ := strings.CutPrefix(object.URLPath, "/api/v4/")
		urlPath, _ := strings.CutPrefix(object.URLPath, "/api/v4")

		for _, field := range object.Fields {
			schemas.Add("", objectName, object.DisplayName, urlPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(schemaManager.FlushSchemas(schemas))
	goutils.MustBeNil(schemaManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}
