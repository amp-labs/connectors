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
		Fields:     connectors.Fields("Id"),
		Since:      time.Now().Add(-600 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading from QuickBooks", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "item",
		Fields:     connectors.Fields("Level", "domain", "Name"),
		Since:      time.Now().Add(-600 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading from QuickBooks", "error", err)
	}

	slog.Info("Reading item..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "term",
		Fields:     connectors.Fields("name"),
		Since:      time.Now().Add(-600 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading from QuickBooks", "error", err)
	}

	slog.Info("Reading term..")
	utils.DumpJSON(res, os.Stdout)

	// Test reading customer with custom fields (requires ?	=enhancedAllCustomFields)
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "customer",
		Fields:     connectors.Fields("Id", "DisplayName"),
		Since:      time.Now().Add(-600 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading from QuickBooks", "error", err)
	}

	slog.Info("Reading customer (with custom fields support)..")
	utils.DumpJSON(res, os.Stdout)

	slog.Info("Read operation completed successfully.")
}
