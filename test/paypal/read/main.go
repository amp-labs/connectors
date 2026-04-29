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
	paypalTest "github.com/amp-labs/connectors/test/paypal"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := paypalTest.GetPayPalConnector(ctx)

	until := time.Now().UTC()
	since := until.AddDate(0, 0, -30)

	// Invoices — no time filter required.
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "invoices",
		Fields:     connectors.Fields("id", "detail", "status", "amount", "primary_recipients"),
	})
	if err != nil {
		utils.Fail("error reading invoices", "error", err)
	}

	slog.Info("Reading invoices...")
	utils.DumpJSON(res, os.Stdout)

	// Disputes — filtered by last-updated time (update_time_after / update_time_before).
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "disputes",
		Fields:     connectors.Fields("dispute_id", "status", "reason", "dispute_amount", "create_time", "update_time"),
		Since:      since,
		Until:      until,
	})
	if err != nil {
		utils.Fail("error reading disputes", "error", err)
	}

	slog.Info("Reading disputes...")
	utils.DumpJSON(res, os.Stdout)

	// Webhook events — filtered by creation time (start_time / end_time).
	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "webhooks-events",
		Fields:     connectors.Fields("id", "event_type", "resource_type", "summary", "create_time"),
		Since:      since,
		Until:      until,
	})
	if err != nil {
		utils.Fail("error reading webhooks-events", "error", err)
	}

	slog.Info("Reading webhooks-events...")
	utils.DumpJSON(res, os.Stdout)
}
