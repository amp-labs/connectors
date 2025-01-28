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

	conn := connTest.GetGoogleConnector(ctx, google.ModuleCalendar)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "settings",
		Fields:     connectors.Fields("id", "value"),
	})
	if err != nil {
		utils.Fail("error reading from Google", "error", err)
	}

	slog.Info("Reading settings..")
	utils.DumpJSON(res, os.Stdout)
}
