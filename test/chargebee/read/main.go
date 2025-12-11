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
	"github.com/amp-labs/connectors/test/chargebee"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := chargebee.GetChargebeeConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("first_name", "last_name"),
	})
	if err != nil {
		utils.Fail("error reading from chargebee", "error", err)
	}

	slog.Info("Reading customers..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "subscriptions",
		Fields:     connectors.Fields("id", "activated_at", "resource_version", "due_invoices_count"),
		Since:      time.Unix(1760342283, 0).UTC(),
	})
	if err != nil {
		utils.Fail("error reading from chargebee", "error", err)
	}

	slog.Info("Reading subscriptions..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "invoices",
		Fields:     connectors.Fields("id"),
		Since:      time.Unix(1760342271, 0).UTC(),
		Until:      time.Unix(1760342283, 0).UTC(),
	})
	if err != nil {
		utils.Fail("error reading from chargebee", "error", err)
	}

	slog.Info("Reading invoices..")
	utils.DumpJSON(res, os.Stdout)
}
