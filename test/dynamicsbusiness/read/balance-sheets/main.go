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
	connTest "github.com/amp-labs/connectors/test/dynamicsbusiness"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetDynamicsBusinessCentralConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "BalanceSheets",
		Fields:     connectors.Fields("id", "display", "lineType"),
		Since:      time.Now(),
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading")
	utils.DumpJSON(res, os.Stdout)
}
