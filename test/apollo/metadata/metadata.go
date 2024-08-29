package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/apollo"
)

func main() {
	ctx := context.Background()

	conn := apollo.GetApolloConnector(ctx, "apollo-creds.json")

	m, err := conn.ListObjectMetadata(ctx, []string{"contact_stages"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
