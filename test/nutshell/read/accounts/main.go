package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/nutshell"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNutshellConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("name"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
