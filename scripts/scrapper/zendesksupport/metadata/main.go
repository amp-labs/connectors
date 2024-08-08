package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/amp-labs/connectors/zendesksupport/metadata"
	"github.com/amp-labs/connectors/zendesksupport/openapi"
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	aliases := api3.NewAliases(map[string]string{
		"search":               "results",
		"ticket_audits":        "audits",
		"satisfaction_reasons": "reasons",
		// problems -> additionalProperties
	})

	objects, err := explorer.GetBasicReadObjects("api/v2", aliases, api3.IdenticalObjectCheck)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	for _, object := range objects {
		for _, field := range object.Fields {
			schemas.Add(aliases.Synonym(object.ObjectName), object.DisplayName, field)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))

	slog.Info("Completed.")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
