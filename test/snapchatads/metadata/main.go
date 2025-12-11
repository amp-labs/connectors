package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/snapchatads"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := snapchatads.GetConnector(ctx)

	_, err := connector.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	m, err := connector.ListObjectMetadata(ctx, []string{"billingcenters", "adaccounts", "members", "roles", "targeting/location/categories_loi"})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
