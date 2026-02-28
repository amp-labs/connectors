package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/granola"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := granola.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"notes"})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
