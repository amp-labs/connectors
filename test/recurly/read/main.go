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
	"github.com/amp-labs/connectors/test/recurly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := recurly.GetRecurlyConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("id", "company", "parent_account_id"),
		Since:      time.Date(2025, 11, 17, 12, 58, 12, 12, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from recurly", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "plans",
		Fields:     connectors.Fields("id", "code", "state"),
	})
	if err != nil {
		utils.Fail("error reading from recurly", "error", err)
	}

	slog.Info("Reading plans..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "items",
		Fields:     connectors.Fields("id", "code", "name"),
	})
	if err != nil {
		utils.Fail("error reading from recurly", "error", err)
	}

	slog.Info("Reading items..")
	utils.DumpJSON(res, os.Stdout)
	slog.Info("Read operation completed successfully.")
}
