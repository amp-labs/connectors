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
		Fields:     connectors.Fields("summary", "description"),
		Deleted:    true,
	})
	if err != nil {
		utils.Fail("error reading from conenctor", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
