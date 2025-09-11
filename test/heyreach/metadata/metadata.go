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

	m, err := conn.ListObjectMetadata(ctx, []string{"campaign/GetAll", "li_account/GetAll", "list/GetAll"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
