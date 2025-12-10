package main

import (
	"context"
	"os"

	connTest "github.com/amp-labs/connectors/test/ciscowebex"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := connTest.GetCiscoWebexConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"people", "roles", "groups"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
