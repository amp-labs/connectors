package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		ObjectName: "invoices",
		Fields:     connectors.Fields("description", "customer_name"),
		Since:      time.Unix(1753116936, 0),
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
