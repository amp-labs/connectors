package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/teamleader"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := teamleader.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"teams"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
