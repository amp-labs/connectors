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

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Opportunity",
		Fields:     connectors.Fields("Id", "Name", "IsDeleted", "SystemModstamp"),
		Since:      timestamp("2025-07-24T13:47:37"),

		// Names must be plural, i.e. "Contacts", "Opportunities", etc.
		AssociatedObjects: []string{"OpportunityContactRoles"},
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Reading..")
	utils.DumpJSON(res, os.Stdout)
}

func timestamp(timeText string) time.Time {
	result, err := time.Parse("2006-01-02T15:04:05", timeText)
	if err != nil {
		utils.Fail("bad timestamp", "error", err)
	}

	return result
}
