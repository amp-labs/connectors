package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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
		ObjectName: "topics",
		Fields:     connectors.Fields("topic_name", "type"),
	})
	if err != nil {
		utils.Fail("error reading from Clari Copilot", "error", err)
	}

	slog.Info("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "calls",
		Fields:     connectors.Fields("id", "title", "type", "status"),
	})
	if err != nil {
		utils.Fail("error reading from Clari Copilot", "error", err)
	}

	slog.Info("Reading calls..")
	utils.DumpJSON(res, os.Stdout)

}
