package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetSalesforceConnector(ctx)

	testscenario.ValidateMetadataContainsRead(ctx, conn, "Organization", sanitizeReadResponse)
}

func sanitizeReadResponse(response map[string]any) map[string]any {
	// every Salesforce response attached attributes object with type and url of a resource.
	// this attribute field will not appear in metadata response, so we shall remove it.
	crucialFields := make(map[string]any)

	for field, v := range response {
		if field != "attributes" {
			crucialFields[field] = v
		}
	}

	return crucialFields
}
