package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/chargeover"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := chargeover.NewConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"customer", "user", "invoice"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
