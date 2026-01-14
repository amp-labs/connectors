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
		ObjectName: "transactions",
		Fields:     connectors.Fields(""),
		Since:      time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		Until:      time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading transactions", "error", err)
	}

	slog.Info("Reading transactions...")
	utils.DumpJSON(res, os.Stdout)
}
