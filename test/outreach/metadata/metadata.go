package main

import (
	"context"
	"fmt"
	"log"

	outreach_test "github.com/amp-labs/connectors/test/outreach"
)

func main() {
	objects := []string{"sequences", "users", "mailings"}

	ctx := context.Background()

	conn := outreach_test.GetOutreachConnector(ctx, "creds.json")

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
