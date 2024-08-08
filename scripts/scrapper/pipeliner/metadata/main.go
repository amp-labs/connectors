package main

import (
	"log"

	"github.com/amp-labs/connectors/pipeliner/metadata"
	"github.com/amp-labs/connectors/pipeliner/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	aliases := api3.NewAliases(map[string]string{})

	objects, err := explorer.GetBasicReadObjects("entities", aliases, api3.DataObjectCheck)
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
