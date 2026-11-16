package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/hunter"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := hunter.GetHunterConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"leads", "leads_lists", "campaigns", "leads_custom_attributes"})
	if err != nil {
		return err
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)

	return nil
}
