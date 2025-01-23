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
		ObjectName:        "payment_intents",
		Fields:            connectors.Fields("currency", "amount"),
		AssociatedObjects: []string{"customer", "application"},
	})
	if err != nil {
		utils.Fail("error reading from Stripe", "error", err)
	}

	slog.Info("Reading payment intents..")
	utils.DumpJSON(res, os.Stdout)
}
