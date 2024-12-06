package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/kit"
)

func main() {
	ctx := context.Background()

	conn := kit.GetKitConnector(ctx)

	// nolint
	m, err := conn.ListObjectMetadata(ctx, []string{"broadcasts", "custom_fields", "forms", "subscribers", "tags", "email_templates", "purchases", "segments", "sequences", "webhooks"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
