package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/dropboxsign"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := dropboxsign.GetDropboxSignConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "template",
		Fields:     connectors.Fields("title", "accounts", "is_creator"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading templates..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "api_app",
		Fields:     connectors.Fields("created_at", "callback_url", "name"),
	})
	if err != nil {
		utils.Fail("error reading from pylon", "error", err)
	}

	slog.Info("Reading api_apps..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
