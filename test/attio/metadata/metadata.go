package main

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/test/attio"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := attio.GetAttioConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"lists", "workspace_members", "tasks", "notes", "companies"})
	if err != nil {
		utils.Fail(err.Error())
	}

	// Print the results.
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
