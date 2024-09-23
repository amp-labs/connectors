package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/attio"
)

func main() {
	ctx := context.Background()

	conn := attio.GetAttioConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"objects", "lists", "self", "workspace_members", "webhooks", "tasks"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results.
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
