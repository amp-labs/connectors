package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/docusign"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := docusign.GetDocusignConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"bulk_send_batch", "bulk_send_lists", "envelopes", "folders", "templates", "users"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
