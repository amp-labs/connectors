package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/common/handy"
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
	must(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride, api3.DataObjectCheck,
	)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()
	registry := handy.NamedLists[string]{}

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)
		}

		for _, field := range object.Fields {
			schemas.Add(object.ObjectName, object.DisplayName, field, nil)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))
	must(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
