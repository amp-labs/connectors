package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/asana"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := asana.GetAsanaConnector(ctx) // nolint

	// nolint
	m, err := conn.ListObjectMetadata(ctx, []string{"projects", "tags", "users", "workspaces"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
