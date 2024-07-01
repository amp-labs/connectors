package main

import (
	"log"

	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/amp-labs/connectors/zendesksupport/metadata"
	"github.com/amp-labs/connectors/zendesksupport/openapi"
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	aliases := handy.NewAliases(map[string]string{
		"search":               "results",
		"ticket_audits":        "audits",
		"satisfaction_reasons": "reasons",
		// problems -> additionalProperties
	})

	objects, err := explorer.GetBasicReadObjects("api/v2", aliases)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	for _, object := range objects {
		for _, field := range object.Fields {
			schemas.Add(aliases.Synonym(object.ObjectName), object.DisplayName, field)
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
