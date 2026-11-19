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

// liveTestReads are the reads a daily accounting export of a store is built from: what the store
// took in, what it sold, the tax it collected, and the cash moved in and out of the drawer.
// Report objects need at least one field: the fields are the report's columns.
//
// In a report, the text fields (dimensions) decide the rows and the amounts (measures) are summed
// per row. Asking sales for report_category and discount_name gives one row per category and
// discount, e.g. "Food" with "10% Off" and "Food" with no discount.
//
//nolint:gochecknoglobals
var liveTestReads = []struct {
	object string
	fields []string
}{
	// Money received, per payment type.
	//   payment_type: how the customer paid (Visa, Cash, or a house account billed later).
	//   gl_code: the GL account set on the payment type in Auphan, often empty.
	//   customer_group, order_id: who a house-account order belongs to, and which order it is.
	//   payment, tips, count: amount paid (card payments include tips), the tips, number of payments.
	{"payments", []string{"payment_type", "gl_code", "customer_group", "order_id", "payment", "tips", "count"}},
	// Items sold, per report category (Food, Beer & Wine, ...).
	//   report_category, gl_code_report_category: the menu category and its GL account.
	//   discount_name, invoice_discount, salerule_name: the item discount, order discount and
	//   automatic promotion (sale rule) applied, empty when none.
	//   gross_sales, net_sales, taxes: sales before discounts, after discounts, and tax on them.
	//   item_discounts, invoice_discounts, salerule_amount: the amount each discount took off.
	{"sales", []string{
		"report_category", "gl_code_report_category", "discount_name", "invoice_discount", "salerule_name",
		"gross_sales", "net_sales", "taxes", "item_discounts", "invoice_discounts", "salerule_amount",
	}},
	// Tax collected, per tax (Sales Tax, ...): name, its GL account, and the total.
	{"tax", []string{"name", "gl_code", "tax_total"}},
	// Cash put into or taken out of the drawer: in or out, the reason typed at the till, GL account, amount.
	{"petty_cash", []string{"petty_cash_type", "reason", "gl_code", "amount"}},
	// The day's order totals, to check that the reports above add up.
	{"invoices", []string{"gross_sales", "net_sales", "total_discounts", "taxes", "total", "tips", "count"}},
	// The store being read.
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
