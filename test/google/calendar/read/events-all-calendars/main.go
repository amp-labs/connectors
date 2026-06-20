package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleCalendarConnector(ctx)

	fields := connectors.Fields("id", "summary", "iCalUID")

	// Read events from every calendar in the user's calendar list, merged and deduped.
	allCalendarsEvents, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "events",
		Fields:     fields,
		Opts:       google.ReadParamsOpts{ReadEventsForAllCalendars: true},
	})
	if err != nil {
		utils.Fail("error reading events for all calendars", "error", err)
	}

	slog.Info("Reading events for all calendars...", "allCalendarsRows", allCalendarsEvents.Rows)

	utils.DumpJSON(allCalendarsEvents, os.Stdout)
}
