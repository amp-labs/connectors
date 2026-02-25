package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetRevenueCatConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"purchases"})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	for objName, objMeta := range m.Result {
		log.Printf("   - %s: %d fields\n", objName, len(objMeta.Fields))
	}

	utils.DumpJSON(m, os.Stdout)
}
