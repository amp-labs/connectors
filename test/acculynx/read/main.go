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
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// liveTestObjects covers the customer-facing in-scope list (Contacts, Jobs,
// Leads, Invoices, Calendar, Custom fields) plus users (implemented but marked
// out-of-scope by the customer — included to validate the implementation
// still works end-to-end).
//
//nolint:gochecknoglobals
var liveTestObjects = []string{
	// Contacts family (in-scope)
	"contacts",
	"contacts/contact-types",
	"contacts/custom-fields",
	"contacts/email-addresses",
	"contacts/phone-numbers",

	// Jobs family (in-scope; jobs/invoices covers the Invoices scope —
	// AccuLynx has no top-level /invoices list endpoint).
	"jobs",
	"jobs/contacts",
	"jobs/custom-fields",
	"jobs/estimates",
	"jobs/history",
	"jobs/invoices",
	"jobs/milestone-history",
	"jobs/representatives",

	// Leads (in-scope). AccuLynx exposes lead-source configuration here;
	// actual lead records live in /jobs at the Lead milestone.
	"company-settings/leads/lead-sources",

	// Calendar family (in-scope)
	"calendars",
	"calendars/appointments",

	// Custom fields catalog (in-scope)
	"company-settings/custom-fields",

	// Users — marked out-of-scope by the customer but the implementation
	// remains; verify it still works.
	"users",
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("Live read test", "count", len(liveTestObjects))

	for _, obj := range liveTestObjects {
		slog.Info("Testing basic read", "object", obj)

		if err := testRead(ctx, conn, obj); err != nil {
			slog.Error(err.Error())
		}
	}

	// Pagination — jobs is the most populated top-level resource.
	slog.Info("Testing pagination for jobs")

	if err := testPagination(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	// Time-based incremental sync — jobs is the only object with a real
	// provider-side ModifiedDate filter; other objects fall back to
	// connector-side Since/Until filtering.
	slog.Info("Testing incremental read for jobs (provider-side ModifiedDate filter)")

	if err := testIncrementalRead(ctx, conn, "jobs"); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *acculynx.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	slog.Info("Read result", "object", objectName, "rows", res.Rows, "nextPage", res.NextPage)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testPagination(ctx context.Context, conn *acculynx.Connector) error {
	objectName := "jobs"
	pageSize := 2

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "jobName", "modifiedDate"),
		PageSize:   pageSize,
	}

	// Read first page.
	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 1: %w", objectName, err)
	}

	slog.Info("Page 1", "object", objectName, "rows", res.Rows, "nextPage", res.NextPage)
	utils.DumpJSON(res, os.Stdout)

	if res.NextPage == "" {
		slog.Warn("No next page found — pagination test may be incomplete", "object", objectName)

		return nil
	}

	// Read second page.
	params.NextPage = res.NextPage

	res2, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 2: %w", objectName, err)
	}

	slog.Info("Page 2", "object", objectName, "rows", res2.Rows)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}

func testIncrementalRead(ctx context.Context, conn *acculynx.Connector, objectName string) error {
	// Read records modified in the last 90 days.
	since := time.Now().AddDate(0, 0, -90)
	until := time.Now()

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "modifiedDate"),
		Since:      since,
		Until:      until,
	}

	slog.Info("Reading with date range",
		"object", objectName,
		"since", since.Format(time.RFC3339),
		"until", until.Format(time.RFC3339))

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s with date range: %w", objectName, err)
	}

	slog.Info("Incremental read result", "object", objectName, "rows", res.Rows)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
