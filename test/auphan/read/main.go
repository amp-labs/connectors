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

// liveTestObjects are the objects The Stand's daily Sage Intacct journal is built from,
// with the columns it uses. Reports need at least one column.
//
//nolint:gochecknoglobals
var liveTestObjects = map[string][]string{
	// Sales journal: tenders and tips per payment type. Also AR (house accounts such as Kaiser).
	"payments": {"date", "payment_type", "gl_code", "payment", "tips", "count"},
	// Sales journal: net sales and discounts per report category.
	"sales": {"date", "report_category", "gl_code_report_category", "net_sales", "total_discounts"},
	// Sales journal: tax collected per tax.
	"tax": {"tax_date", "name", "gl_code", "tax_total"},
	// Sales journal: petty cash in and out.
	"petty_cash": {"tracking_date", "petty_cash_type", "gl_code", "amount"},
	// AR: house account customers.
	"customers": {"customer_id", "company", "invoices", "sales"},
	// Store list, one journal per store.
	"stores": {"store_id", "name", "timezone"},
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := auphan.GetAuphanConnector(ctx)

	// One business day, the shape of the daily journal run.
	since := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	until := since.Add(24*time.Hour - time.Second)

	for objectName, fields := range liveTestObjects {
		slog.Info("Testing read", "object", objectName)

		if err := testRead(ctx, conn, objectName, fields, since, until); err != nil {
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

	slog.Info("Read result", "object", objectName, "rows", res.Rows, "done", res.Done)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
