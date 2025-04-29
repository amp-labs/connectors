package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/fireflies"
)

func main() {
	ctx := context.Background()

	conn := fireflies.GetFirefliesConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"user", "transcript", "bite"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
