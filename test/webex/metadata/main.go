package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/webex"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := connTest.GetWebexConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"people", "roles", "groups"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
