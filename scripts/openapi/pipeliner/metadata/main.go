package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/pipeliner/metadata"
	"github.com/amp-labs/connectors/providers/pipeliner/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"*/batch-modify",
		"*/batch-delete",
		"/entities/Accounts/merge",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"AccountKPIs":   "Account KPIs",
		"ActivityKPIs":  "Activity KPIs",
		"ApiAccesses":   "API Accesses",
		"ContactKPIs":   "Contact KPIs",
		"LeadOpptyKPIs": "Lead Oppty KPIs",
		"ProjectKPIs":   "Project KPIs",
		"QuoteKPIs":     "Quote KPIs",
		"Webresources":  "Web Resources",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		))
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride, api3.DataObjectLocator,
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

		urlPath, _ := strings.CutPrefix(object.URLPath, "/entities")

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, urlPath, object.ResponseKey,
				staticschema.FieldMetadataMapV1{
					field.Name: field.Name,
				}, nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}
