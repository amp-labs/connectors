package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/attio"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := attio.GetAttioConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"lists", "workspace_members", "tasks", "notes", "companies", "deals", "people", "users", "workspaces"})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results.
	utils.DumpJSON(m, os.Stdout)
}
