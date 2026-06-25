package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "checkout/sessions",
		Fields: connectors.Fields(
			"object",
			// "line_items",
			"$['line_items']['data'][*]['currency']",
			"$['line_items']['data'][*]['id']",
			"$['line_items']['has_more']",
			"$['line_items']['url']",
		),
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
