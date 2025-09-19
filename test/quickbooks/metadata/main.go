package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/quickbooks"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := quickbooks.GetQuickBooksConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"account", "budget"})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
