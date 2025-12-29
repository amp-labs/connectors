package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	msTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := msTest.GetSalesloftConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "notes",
		Fields:     connectors.Fields("content"),
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
