package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/pinterest"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := pinterest.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"pins", "boards", "media", "ad_accounts", "catalogs"})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
