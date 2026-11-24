package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/front"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// Values of contacts and a conversation that exist in the test account.
// The phone belongs to the contact crd_5hfk3q4 ("Search Phone Test").
const (
	contactID          = "crd_4mhsx24"
	contactEmail       = "dwightb@limitlesspaper.com"
	contactPhone       = "+15555550100"
	conversationID     = "cnv_1alzn15o"
	conversationStatus = "archived"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetFrontConnector(ctx)

	slog.Info("Searching contact by id...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "name", "handles"),
		Filter:     equals("id", contactID),
	})

	slog.Info("Searching contact by email...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "name", "handles"),
		Filter:     equals("email", contactEmail),
	})

	slog.Info("Searching contact by phone...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "name", "handles"),
		Filter:     equals("phone", contactPhone),
	})

	slog.Info("Searching conversation by id...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "conversations",
		Fields:     connectors.Fields("id", "subject", "status"),
		Filter:     equals("id", conversationID),
	})

	slog.Info("Searching conversations by status...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "conversations",
		Fields:     connectors.Fields("id", "subject", "status"),
		Filter:     equals("status", conversationStatus),
		Limit:      1,
	})
}

func equals(fieldName string, value any) connectors.SearchFilter {
	return connectors.SearchFilter{
		FieldFilters: []common.FieldFilter{
			{
				FieldName: fieldName,
				Operator:  common.FilterOperatorEQ,
				Value:     value,
			},
		},
	}
}
