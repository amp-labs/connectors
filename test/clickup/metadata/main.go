package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/clickup"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := clickup.GetClickupConnector(ctx) // nolint

	// nolint
	m, err := conn.ListObjectMetadata(ctx, []string{"team"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
