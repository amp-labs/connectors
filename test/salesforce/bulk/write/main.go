package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Pre-requisites:
// (1)	Opportunity object must have ExternalID.
// 		This means you need to add custom field and mark it as external ID.
//		Steps:
//			* Login to Salesforce
//			* Open "Object Manager"
//			* Search for "Opportunities"
//			* Tab - "Fields & Relationships"
//			* Click on "New", select "Text", click "Next
//			* Give field "external_id" name (within API it will be known as "external_id__c")
//			* Tick External ID box!
//

var tests = testutils.ParallelRunners[*salesforce.Connector]{
	{
		FilePath:  "opportunities.csv",
		TestTitle: "Testing Bulk Write",
		Function:  testBulkWriteOpportunity,
	},
	{
		FilePath:  "opportunities.csv",
		TestTitle: "Testing Success Results",
		Function:  testGetJobResultsForFile,
	},
	{
		// Failure due to invalid field: deserialize timestamp (CloseDate field)
		FilePath:  "opportunities-partial-failure.csv",
		TestTitle: "Testing Partial Failure",
		Function:  testGetJobResultsForFile,
	},
	{
		// Failure due to invalid field (StageName field)
		FilePath:  "opportunities-complete-failure.csv",
		TestTitle: "Testing Complete Failure",
		Function:  testGetJobResultsForFile,
	},
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	defer utils.Close(conn)

	logs := tests.Run(ctx, conn)

	for _, log := range logs {
		fmt.Println(log)
	}
}
