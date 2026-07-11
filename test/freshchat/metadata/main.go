package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/freshchat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := freshchat.NewConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"users", "agents", "groups"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
