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
	"github.com/amp-labs/connectors/providers/aircall"
	testAircall "github.com/amp-labs/connectors/test/aircall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testAircall.GetAircallConnector(ctx)

	slog.Info("Testing basic read for calls")
	if err := testRead(ctx, conn, "calls"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for users")
	if err := testRead(ctx, conn, "users"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for teams")
	if err := testRead(ctx, conn, "teams"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for tags")
	if err := testRead(ctx, conn, "tags"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for contacts")
	if err := testRead(ctx, conn, "contacts"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for numbers")
	if err := testRead(ctx, conn, "numbers"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing time-based filtering/incremental sync for calls")
	if err := testIncrementalRead(ctx, conn, "calls"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing pagination for tags")
	if err := testPagination(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing time-based filtering/incremental sync for users")
	if err := testUsersFiltering(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing pagination for contacts")
	if err := testContactsPagination(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing time-based filtering/incremental sync for contacts")
	if err := testContactsFiltering(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func testContactsPagination(ctx context.Context, conn *aircall.Connector) error {
	objectName := "contacts"
	pageSize := 2

	slog.Info("Testing pagination for contacts", "object", objectName, "pageSize", pageSize)

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "first_name", "last_name"),
		PageSize:   pageSize,
	}

	// Read first page
	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 1: %w", objectName, err)
	}

	slog.Info("Page 1 results", "rows", res.Rows, "nextPage", res.NextPage)
	utils.DumpJSON(res, os.Stdout)

	if res.NextPage == "" {
		slog.Warn("No next page found for contacts. Pagination test might be incomplete.")
		return nil
	}

	// Read second page
	params.NextPage = res.NextPage
	res2, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 2: %w", objectName, err)
	}

	slog.Info("Page 2 results", "rows", res2.Rows)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}

func testContactsFiltering(ctx context.Context, conn *aircall.Connector) error {
	// Filter contacts created in the last 7 days
	since := time.Now().Add(-7 * 24 * time.Hour)
	until := time.Now()

	slog.Info("Testing contacts filtering (recent)", "since", since, "until", until)
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "first_name", "last_name", "created_at"),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading contacts with range: %w", err)
	}
	slog.Info("Recent contacts results", "rows", res.Rows)

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testRead(ctx context.Context, conn *aircall.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	// Print the results.
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testIncrementalRead(ctx context.Context, conn *aircall.Connector, objectName string) error {
	// Test incremental sync: Get records from the last 7 days
	since := time.Now().AddDate(0, 0, -7)
	until := time.Now()

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "created_at"),
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

	slog.Info("Incremental read results",
		"object", objectName,
		"rows", res.Rows,
		"done", res.Done)

	// Print the results.
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testPagination(ctx context.Context, conn *aircall.Connector) error {
	// Test pagination with tags (assuming we have multiple tags)
	objectName := "tags"
	pageSize := 2

	slog.Info("Testing pagination", "object", objectName, "pageSize", pageSize)

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "name"),
		PageSize:   pageSize,
	}

	// Read first page
	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s page 1: %w", objectName, err)
	}

	slog.Info("Page 1 results", "rows", res.Rows, "nextPage", res.NextPage)
	utils.DumpJSON(res, os.Stdout)

	if res.NextPage == "" {
		slog.Warn("No next page found. Pagination test might be incomplete if there are fewer items than PageSize.")
		return nil
	}

	// Read second page
	params.NextPage = res.NextPage
	res2, err := conn.Read(ctx, params)
	if err != nil {
	}

	slog.Info("Page 2 results", "rows", res2.Rows)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}

func testUsersFiltering(ctx context.Context, conn *aircall.Connector) error {
	// User created at 2025-11-21T16:58:07Z
	// Test 1: Range including the user
	since := time.Date(2025, 11, 20, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 11, 22, 0, 0, 0, 0, time.UTC)

	slog.Info("Testing users filtering (should include user)", "since", since, "until", until)
	params := common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name", "created_at"),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading users with range 1: %w", err)
	}
	slog.Info("Range 1 results", "rows", res.Rows)
	utils.DumpJSON(res, os.Stdout)

	// Test 2: Range excluding the user (past)
	sincePast := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	untilPast := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	slog.Info("Testing users filtering (should exclude user)", "since", sincePast, "until", untilPast)
	params.Since = sincePast
	params.Until = untilPast

	res2, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading users with range 2: %w", err)
	}
	slog.Info("Range 2 results", "rows", res2.Rows)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}
