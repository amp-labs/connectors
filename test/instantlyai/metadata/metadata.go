package main

import (
	"context"
	"fmt"
	"log"

	instantlyai "github.com/amp-labs/connectors/test/instantlyai"
)

func main() {
	ctx := context.Background()

	conn := instantlyai.GetInstantlyAIConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"accounts", "campaigns", "emails", "lead-lists", "inbox-placement-tests", "background-jobs", "custom-tags", "block-lists-entries", "lead-labels", "workspace-group-members", "workspace-members", "leads/list", "campaigns/analytics", "campaigns/analytics/daily", "campaigns/analytics/steps", "inbox-placement-tests/email-service-provider-options", "audit-logs"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
