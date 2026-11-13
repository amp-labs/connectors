package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/auphan"
	"github.com/amp-labs/connectors/test/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

// liveTestReads are the reads The Stand's daily Sage Intacct journals are built from, with the
// columns each uses. Reports need at least one column. sales is read twice, grouped two ways.
//
//nolint:gochecknoglobals
var liveTestReads = []struct {
	object string
	fields []string
}{
	// Sales journal debits: receipts per payment type. Card payments include their tips.
	// AR: house-account payment types (e.g. Kaiser), one row per order.
	{"payments", []string{"payment_type", "gl_code", "customer_group", "order_id", "payment", "tips", "count"}},
	// Sales journal credits: sales before discounts per report category (Food, Beer & Wine, ...).
	// Sales journal debits: discounts and sale rules (coupons, loyalty, employee meals).
	// Auphan groups by every dimension asked for, so rows are per category and discount; amounts add up.
	{"sales", []string{
		"report_category", "gl_code_report_category", "discount_name", "invoice_discount", "salerule_name",
		"gross_sales", "net_sales", "taxes", "item_discounts", "invoice_discounts", "salerule_amount",
	}},
	// Sales journal credit: sales tax collected.
	{"tax", []string{"name", "gl_code", "tax_total"}},
	// Sales journal: petty cash in and out of the drawer.
	{"petty_cash", []string{"petty_cash_type", "reason", "gl_code", "amount"}},
	// DSS tie-out: the day's invoice totals.
	{"invoices", []string{"gross_sales", "net_sales", "total_discounts", "taxes", "total", "tips", "count"}},
	// One journal per store.
	{"stores", []string{"store_id", "name", "timezone"}},
}

// Reads one business day, the shape of the daily journal run: 2026-11-05, or the date given
// as the first argument (YYYY-MM-DD).
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := auphan.GetAuphanConnector(ctx)

	since := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)

	if len(os.Args) > 1 {
		day, err := time.Parse(time.DateOnly, os.Args[1])
		if err != nil {
			utils.Fail("date must be YYYY-MM-DD", "error", err)
		}

		since = day
	}

	until := since.Add(24*time.Hour - time.Second)

	for _, read := range liveTestReads {
		slog.Info("Testing read", "object", read.object, "fields", read.fields)

		if err := testRead(ctx, conn, read.object, read.fields, since, until); err != nil {
			slog.Error(err.Error())
		}
	}
}

func testRead(ctx context.Context, conn *ap.Connector, objectName string, fields []string,
	since, until time.Time,
) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      since,
		Until:      until,
	})
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	slog.Info("Read result", "object", objectName, "fields", fields, "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
