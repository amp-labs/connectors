package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/xero"
)

func main() {
	ctx := context.Background()
	connector := xero.GetXeroConnector(ctx)

	_, err := connector.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	m, err := connector.ListObjectMetadata(ctx, []string{"contacts", "accounts", "Budgets"})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
