package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/heyreach"
)

func main() {
	ctx := context.Background()

	conn := heyreach.GetHeyreachConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"campaign", "li_account", "list"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
