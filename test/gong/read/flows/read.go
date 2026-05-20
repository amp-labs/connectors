package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGongConnector(ctx)

	slog.Info("Reading flows aggregated across all users..")
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "flows",
		Fields:     connectors.Fields("id", "name", "visibility"),
	})
	if err != nil {
		utils.Fail("error reading from Gong", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading flows with users association..")
	resWithAssoc, err := conn.Read(ctx, common.ReadParams{
		ObjectName:        "flows",
		Fields:            connectors.Fields("id", "name", "visibility"),
		AssociatedObjects: []string{"users"},
	})
	if err != nil {
		utils.Fail("error reading from Gong with associations", "error", err)
	}

	utils.DumpJSON(resWithAssoc, os.Stdout)
}
