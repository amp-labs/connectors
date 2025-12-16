package main

import (
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/cloudtalk/metadata"
	"github.com/amp-labs/connectors/providers/cloudtalk/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Initial ignore list, can be refined based on output.
//
//nolint:gochecknoglobals
var ignoreEndpoints = []string{
	"/agents/add.json",
	"/agents/edit/{agentId}.json",
	"/agents/delete/{agentId}.json",
	"/groups/add.json",
	"/groups/delete/{agent_id}.json",
	// Add more non-list endpoints here to reduce noise.

}

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)

	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, nil, api3.DataObjectLocator,
	)

	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	titleCaser := cases.Title(language.English)

	for _, object := range readObjects {
		// CloudTalk list endpoints usually end in /index.json.
		// We filter out any paths that do not match this pattern (except for /contacts.json which is a bulk endpoint)
		// to avoid processing non-resource endpoints.
		if !strings.HasSuffix(object.URLPath, "/index.json") &&
			!strings.HasSuffix(object.URLPath, "/contacts.json") { // /bulk/contacts.json
			continue
		}

		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			// Clean up object name to be more friendly.
			// e.g. /agents/index.json -> agents
			objName := strings.TrimPrefix(object.ObjectName, "/")
			objName = strings.TrimSuffix(objName, "/index.json")
			objName = strings.TrimSuffix(objName, ".json")

			displayName := titleCaser.String(objName)

			schemas.Add("", objName, displayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}
