package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/orb"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := orb.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"customers", "subscriptions", "invoices", "metrics"})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
