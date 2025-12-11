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
	"github.com/amp-labs/connectors/test/linear"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := linear.GetLinearConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "issues",
		Fields:     connectors.Fields("id", "title", "state"),
		Since:      time.Date(2025, 06, 23, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from Linear", "error", err)
	}

	slog.Info("Reading issues..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "projectStatuses",
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from Linear", "error", err)
	}

	slog.Info("Reading project statuses..")
	utils.DumpJSON(res, os.Stdout)

}
