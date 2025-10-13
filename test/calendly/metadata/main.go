package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/calendly"
)

func main() {
	ctx := context.Background()

	conn := calendly.GetCalendlyConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"activity_log_entries", "scheduled_events", "tests"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
