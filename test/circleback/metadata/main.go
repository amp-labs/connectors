package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/circleback"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := circleback.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"meetings", "action-items", "tag", "people", "companies",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
