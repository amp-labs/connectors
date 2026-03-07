package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/expensify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "policy",
		Fields:     connectors.Fields("id", "name", "created"),
	})
	if err != nil {
		utils.Fail("error reading from Expensify", "error", err)
	}

	slog.Info("Reading policies..")
	utils.DumpJSON(res, os.Stdout)

}
