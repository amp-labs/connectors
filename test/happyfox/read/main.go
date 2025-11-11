package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/happyfox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := happyfox.GetHappyFoxConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "profiles",
		Fields: connectors.Fields(
			"name",
		),
	})
	if err != nil {
		utils.Fail("error reading from Instantly", "error", err)
	}

	slog.Info("Reading profiles..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "departments",
		Fields: connectors.Fields(
			"name",
		),
	})
	if err != nil {
		utils.Fail("error reading from Instantly", "error", err)
	}

	slog.Info("Reading departments..")
	utils.DumpJSON(res, os.Stdout)
}
