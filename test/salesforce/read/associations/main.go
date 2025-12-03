package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const TimeoutSeconds = 30

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	ctx, done = context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	// Test Account associations
	fmt.Println("Testing Account associations (Account -> Contacts, Opportunities)...")
	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     connectors.Fields("Id", "Name", "BillingCity", "IsDeleted", "SystemModstamp"),
		Since:      timestamp("2024-08-28T13:47:37"),

		// Names must be plural, i.e. "Contacts", "Opportunities", etc.
		AssociatedObjects: []string{"Contacts", "Opportunities"},
	})
	if err != nil {
		utils.Fail("error reading Account associations", "error", err)
	}

	fmt.Println("Account associations result:")
	utils.DumpJSON(res, os.Stdout)

	// Test Opportunity associations
	fmt.Println("\nTesting Opportunity associations (Opportunity -> Account, Contacts)...")
	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Opportunity",
		Fields:     connectors.Fields("Id", "Name", "Amount", "StageName", "CloseDate"),
		// Test both parent relationship (account) and junction relationship (contacts)
		AssociatedObjects: []string{"account", "contacts"},
	})
	if err != nil {
		utils.Fail("error reading Opportunity associations", "error", err)
	}

	fmt.Println("Opportunity associations result:")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
