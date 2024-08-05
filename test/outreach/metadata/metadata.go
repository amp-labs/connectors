package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/outreach"
)

func main() {
	objects := []string{"sequences"}

	ctx := context.Background()

	conn := outreach.GetOutreachConnector(ctx, "creds.json")

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
