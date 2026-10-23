package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/sybill"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := sybill.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"conversations", "accounts", "deals"})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
