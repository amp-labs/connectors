package main

import (
	"log"

	"github.com/amp-labs/connectors/pipeliner/metadata"
	"github.com/amp-labs/connectors/pipeliner/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
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
		// none
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
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
