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
	"github.com/amp-labs/connectors/test/claricopilot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := claricopilot.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name"),
	})
	if err != nil {
		utils.Fail("error reading from Clari Copilot", "error", err)
	}

	slog.Info("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "calls",
		Fields:     connectors.Fields("id", "title", "type", "status"),
		Since:      time.Date(2025, 06, 20, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Clari Copilot", "error", err)
	}

	slog.Info("Reading calls..")
	utils.DumpJSON(res, os.Stdout)
}
