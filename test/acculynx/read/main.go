package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
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

	// Regression: platform sends Since set + Until zero on backfills. /jobs
	// and /jobs/{id}/history return HTTP 400 when only one of startDate /
	// endDate is sent. Exercise the exact param shape here so a regression
	// surfaces in live testing rather than in production.
	for _, obj := range []string{"jobs", "jobs/history"} {
		slog.Info("Testing backfill-shape incremental read (Since set, Until zero)", "object", obj)

		if err := testBackfillShapeRead(ctx, conn, obj); err != nil {
			slog.Error(err.Error())
		}
	}

	slog.Info("Testing jobs read from a saved cursor past the 100000 pageStartIndex limit")

	if err := testJobsCursorPastLimit(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

// testJobsCursorPastLimit starts a jobs read from a saved next-page URL at
// pageStartIndex=100000, like the one a customer's backfill got stuck on.
// AccuLynx answers that URL with 400 "Requested pageStartIndex is out of
// range.", so the connector must restart the list sorted by modifiedDate and
// still return every job a normal read returns.
func testJobsCursorPastLimit(ctx context.Context, conn *acculynx.Connector) error {
	params := common.ReadParams{
		ObjectName: "jobs",
		Fields:     connectors.Fields("id"),
		Since:      time.Now().AddDate(-10, 0, 0),
	}

	first, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading jobs page 1: %w", err)
	}

	if first.NextPage == "" {
		slog.Warn("Only one page of jobs, so there is no next-page URL to test with")

		return nil
	}

	cursor, err := url.Parse(first.NextPage.String())
	if err != nil {
		return err
	}

	query := cursor.Query()
	query.Set("pageStartIndex", "100000")
	cursor.RawQuery = query.Encode()

	want, err := readJobIDs(ctx, conn, params)
	if err != nil {
		return fmt.Errorf("error reading jobs: %w", err)
	}

	stuck := params
	stuck.NextPage = common.NextPageToken(cursor.String())

	got, err := readJobIDs(ctx, conn, stuck)
	if err != nil {
		return fmt.Errorf("error reading jobs from %s: %w", cursor, err)
	}

	missing := 0

	for id := range want {
		if !got[id] {
			missing++
		}
	}

	if missing > 0 {
		return fmt.Errorf("read from the saved cursor missed %d of %d jobs", missing, len(want)) //nolint:err113
	}

	slog.Info("Saved cursor past the limit read every job", "jobs", len(want))

	return nil
}

// readJobIDs reads every page and returns the IDs of the jobs it got.
func readJobIDs(ctx context.Context, conn *acculynx.Connector, params common.ReadParams) (map[string]bool, error) {
	ids := map[string]bool{}

	for {
		res, err := conn.Read(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, row := range res.Data {
			ids[fmt.Sprint(row.Fields["id"])] = true
		}

		if res.Done || res.NextPage == "" {
			return ids, nil
		}

		params.NextPage = res.NextPage
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

// testBackfillShapeRead mirrors the platform's backfill param shape: Since set,
// Until zero. Before the date-filter fix, this combination produced HTTP 400
// from AccuLynx ("Start Date and End Date do not have the same format").
func testBackfillShapeRead(ctx context.Context, conn *acculynx.Connector, objectName string) error {
	since := time.Now().AddDate(0, 0, -15)

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
		Since:      since,
		// Until intentionally left as zero value.
	}

	slog.Info("Reading with backfill shape",
		"object", objectName,
		"since", since.Format(time.RFC3339),
		"until", "zero")

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s with backfill shape: %w", objectName, err)
	}

	slog.Info("Backfill-shape read result", "object", objectName, "rows", res.Rows)
	utils.DumpJSON(res, os.Stdout)

	return nil
}
