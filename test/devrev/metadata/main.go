package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/devrev"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := devrev.GetConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"commands", "articles",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
