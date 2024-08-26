package main

import (
	"log"

	"github.com/amp-labs/connectors/providers/pipeliner/metadata"
	"github.com/amp-labs/connectors/providers/pipeliner/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/iancoleman/strcase"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"*/batch-modify",
		"*/batch-delete",
		"/entities/Accounts/merge",
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		// none
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
			func(displayName string) string {
				// Camel case changed to space delimited.
				return strcase.ToDelimited(displayName, ' ')
			},
			api3.CapitalizeFirstLetterEveryWord,
		))
	must(err)

	objects, err := explorer.GetBasicReadObjects(
		ignoreEndpoints, objectEndpoints, displayNameOverride, api3.DataObjectCheck,
	)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	for _, object := range objects {
		for _, field := range object.Fields {
			schemas.Add(object.ObjectName, object.DisplayName, field, nil)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))

	log.Println("Completed.")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
