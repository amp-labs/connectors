package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	braintreeTest "github.com/amp-labs/connectors/test/braintree"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := braintreeTest.GetBraintreeConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "disputes",
		Fields:     connectors.Fields(""),
	})
	if err != nil {
		utils.Fail("error reading disputes", "error", err)
	}

	slog.Info("Reading disputes...")
	utils.DumpJSON(res, os.Stdout)
}
