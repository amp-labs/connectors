package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/xero"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := xero.GetXeroConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("code", "name", "Type"),
	})

	if err != nil {
		utils.Fail("error reading from Xero", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
