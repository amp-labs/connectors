package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/greenhouse"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGreenhouseConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "candidates",
		Fields:     connectors.Fields("id", "first_name", "last_name", "updated_at"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	slog.Info("Reading candidates...")
	utils.DumpJSON(res, os.Stdout)
}
