package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/apollo"
)

func main() {
	ctx := context.Background()

	conn := apollo.GetApolloConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"opportunities", "contact_stages", "email_accounts", "typed_custom_fields", "opportunity_stages", "users", "deals"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
