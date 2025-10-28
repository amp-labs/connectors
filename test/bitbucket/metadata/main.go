package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/bitbucket"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := bitbucket.GetConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"workspaces", "repositories"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
