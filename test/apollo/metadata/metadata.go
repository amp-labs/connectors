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

	m, err := conn.ListObjectMetadata(ctx, []string{"opportunities", "contact_stages", "email_accounts", "typed_custom_fields", "opportunity_stages", "users"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
