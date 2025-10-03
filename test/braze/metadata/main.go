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

	objectNames := []string{"catalogs", "campaigns", "templates/email"}

	m, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
