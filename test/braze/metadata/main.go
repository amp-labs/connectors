package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/braze"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := braze.NewBrazeConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"catalogs", "campaigns", "templates/email"})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
