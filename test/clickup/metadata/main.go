package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/clickup"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := clickup.GetClickupConnector(ctx) // nolint

	// nolint
	result, err := conn.ListObjectMetadata(ctx, []string{"channels", "groups", "tasks", "teams"})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	utils.DumpJSON(result, os.Stdout)
}
