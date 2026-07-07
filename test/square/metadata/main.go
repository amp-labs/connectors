package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/square"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := square.GetSquareConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"customers",
		"locations",
		"payments",
		"catalog",
		"merchants",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
