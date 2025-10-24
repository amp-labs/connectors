package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/chargebee"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := chargebee.GetChargebeeConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"subscriptions", "events", "items"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
