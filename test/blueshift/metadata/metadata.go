package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/blueshift"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := blueshift.GetBlueshiftConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"campaigns", "tag_contexts/list", "external_fetches"})

	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
