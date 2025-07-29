package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/flatfile"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := flatfile.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "events",
		Fields:     connectors.Fields("id", "subject", "mode"),
		Since:      time.Date(2025, 07, 0, 0, 0, 0, 0, time.UTC),
	})

	if err != nil {
		utils.Fail("error reading from flatfile", "error", err)
	}

	slog.Info("Reading events..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name", "email"),
	})

	if err != nil {
		utils.Fail("error reading from flatfile", "error", err)
	}

	slog.Info("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "apps",
		Fields:     connectors.Fields("id", "name", "type"),
	})

	if err != nil {
		utils.Fail("error reading from flatfile", "error", err)
	}

	slog.Info("Reading apps..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "environments",
		Fields:     connectors.Fields("id", "accountId", "name"),
	})

	if err != nil {
		utils.Fail("error reading from flatfile", "error", err)
	}
	slog.Info("Reading environments..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
