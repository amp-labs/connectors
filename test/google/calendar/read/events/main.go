package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "events",
		// JSONPath names resolve values from the nested start/end/source objects.
		Fields: connectors.Fields("summary", "$['start']['dateTime']", "$['start']['date']",
			"$['end']['dateTime']", "$['end']['timeZone']", "$['source']['title']", "$['extendedProperties']['private']['crmId']"),
		Deleted: true,
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	slog.Info("Reading events from primary calendars", "read events", res.Rows)
	utils.DumpJSON(res, os.Stdout)
}
