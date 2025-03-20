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
	connTest "github.com/amp-labs/connectors/test/aha"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAhaConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "audits",
		Fields:     connectors.Fields("id", "created_at", "user"),
		Since:      time.Now().AddDate(-1, 0, 0),
	})
	if err != nil {
		utils.Fail("error reading from Aha", "error", err)
	}

	slog.Info("Reading audits..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "screen_definitions",
		Fields:     connectors.Fields("id", "name", "custom_field_definitions", "screenable_type"),
	})
	if err != nil {
		utils.Fail("error reading from Aha", "error", err)
	}

	slog.Info("Reading screen_definitions..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "custom_field_definitions",
		Fields:     connectors.Fields("id", "name", "key", "type"),
	})
	if err != nil {
		utils.Fail("error reading from Aha", "error", err)
	}

	slog.Info("Reading custom_field_definitions..")
	utils.DumpJSON(res, os.Stdout)
}
