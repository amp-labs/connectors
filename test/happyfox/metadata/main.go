package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/happyfox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := happyfox.GetHappyFoxConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"agents", "profiles"})
	if err != nil {
		return err
	}

	// Print the results
	utils.DumpJSON(m.Result, os.Stdout)

	return nil
}
