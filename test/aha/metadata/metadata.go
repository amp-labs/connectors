package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/aha"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := aha.GetAhaConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"audits", "idea_organizations", "me/assigned"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
