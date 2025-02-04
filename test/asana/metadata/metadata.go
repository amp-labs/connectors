package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/asana"
)

func main() {
	ctx := context.Background()

	conn := asana.GetAsanaConnector(ctx) // nolint

	// nolint
	m, err := conn.ListObjectMetadata(ctx, []string{"projects", "tags", "users", "workspaces"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
