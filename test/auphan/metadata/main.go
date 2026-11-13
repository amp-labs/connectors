package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := auphan.GetAuphanConnector(ctx)

	utils.SetupLogging()

	res, err := conn.ListObjectMetadata(ctx, []string{"payments", "sales", "stores", "tax", "customers"})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

}
