package main

import (
	"context"
	"fmt"
	"log"

	instantlyAI "github.com/amp-labs/connectors/test/instantlyAI"
)

func main() {
	ctx := context.Background()

	conn := instantlyAI.GetInstantlyAIConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"accounts", "campaigns", "emails", "lead-lists", "inbox-placement-tests", "inbox-placement-analytics", "inbox-placement-reports", "api-keys", "background-jobs", "custom-tags", "block-lists-entries", "lead-labels", "workspace-group-members", "workspace-members", "subsequences"})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
