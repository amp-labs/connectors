package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/clickup"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetClickupConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "team",
		Fields:     connectors.Fields("id", "name", "color", "avatar"),
	})
	if err != nil {
		utils.Fail("error reading from Clickup", "error", err)
	}

	slog.Info("Reading team..")
	utils.DumpJSON(res, os.Stdout)
}
