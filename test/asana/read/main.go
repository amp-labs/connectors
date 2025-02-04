package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/asana"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAsanaConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "workspaces",
		Fields:     connectors.Fields("gid"),
	})
	if err != nil {
		utils.Fail("error reading from Asana", "error", err)
	}

	slog.Info("Reading projects..")
	utils.DumpJSON(res, os.Stdout)
}
