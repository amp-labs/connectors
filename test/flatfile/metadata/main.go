package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/flatfile"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := flatfile.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"apps", "users", "spaces", "jobs"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)

	return nil
}
