package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/g2"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := g2.NewConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"buyer_intent", "categories", "discussions", "products"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
