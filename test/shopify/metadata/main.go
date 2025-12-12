package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/shopify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := shopify.GetShopifyConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"products", "orders", "customers"})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	for objName, objMeta := range m.Result {
		log.Printf("   - %s: %d fields\n", objName, len(objMeta.Fields))
	}

	utils.DumpJSON(m, os.Stdout)
}
