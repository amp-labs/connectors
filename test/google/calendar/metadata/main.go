package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "calendarList"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleConnector(ctx, google.ModuleCalendar)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Google", "error", err)
	}

	slog.Info("Metadata " + objectName + "..")
	utils.DumpJSON(metadata, os.Stdout)
}
