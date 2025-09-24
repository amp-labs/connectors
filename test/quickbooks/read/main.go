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
	"github.com/amp-labs/connectors/test/quickbooks"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := quickbooks.GetQuickBooksConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "account",
		Fields:     connectors.Fields("Name", "domain", "Classification"),
		Since:      time.Now().Add(-400 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading from Xero", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
