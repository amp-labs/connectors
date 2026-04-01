package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetFastSpringConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"accounts",
		"orders",
		"products",
		"subscriptions",
		"events-processed",
		"events-unprocessed",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
