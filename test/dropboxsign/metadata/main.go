package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/dropboxsign"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := dropboxsign.GetDropboxSignConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"template", "api_app"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)

	return nil
}
