package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/recurly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := recurly.GetRecurlyConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"accounts", "invoices", "subscriptions"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)

	return nil
}
