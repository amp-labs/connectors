package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := revenuecat.GetRevenueCatConnector(ctx)

	meta, err := conn.ListObjectMetadata(ctx, []string{"customers"})
	if err != nil {
		log.Fatalf("ListObjectMetadata error: %v", err)
	}

	utils.DumpJSON(meta, os.Stdout)
}
