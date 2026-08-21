package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/monaco"
	"github.com/amp-labs/connectors/test/utils"
)

// objects is every collection declared in providers/monaco/metadata/schemas.json.
//
//nolint:gochecknoglobals
var objects = []string{
	"accounts",
	"audiences",
	"campaigns",
	"contacts",
	"meetings",
	"opportunities",
	"sequenceTemplates",
	"sequences",
	"tags",
	"tasks",
	"users",
}

func main() {
	ctx := context.Background()

	conn := connTest.GetMonacoConnector(ctx)

	// Schemas are static, so this resolves without touching the network -- it
	// runs green even without a valid API key.
	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	for objName, objMeta := range m.Result {
		log.Printf("   - %s (%s): %d fields\n", objName, objMeta.DisplayName, len(objMeta.Fields))
	}

	for objName, objErr := range m.Errors {
		log.Printf("   ! %s: %v\n", objName, objErr)
	}

	utils.DumpJSON(m, os.Stdout)
}
