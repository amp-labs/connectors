package main

import (
	"context"
	"fmt"

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

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
