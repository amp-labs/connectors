package main

import (
	"context"
	"os/signal"
	"strings"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetPipelinerConnector(ctx)

	testscenario.ValidateMetadataContainsRead(ctx, conn, "Leads", sanitizeReadResponse)
}

func sanitizeReadResponse(response map[string]any) map[string]any {
	// Pipeliner has some extra fields attached starting with `cf_` prefix.
	// Example Leads:
	//	cf_other_lead_source
	//	cf_lead_source1
	//	cf_lead_source1_id
	crucialFields := make(map[string]any)

	for field, v := range response {
		if !strings.HasPrefix(field, "cf_") {
			crucialFields[field] = v
		}
	}

	return crucialFields
}
