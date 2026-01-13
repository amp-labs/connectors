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
	"github.com/amp-labs/connectors/providers/cloudtalk"
	testCloudTalk "github.com/amp-labs/connectors/test/cloudtalk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testCloudTalk.GetCloudTalkConnector(ctx)

	slog.Info("Testing basic read for contacts")

	if err := testRead(ctx, conn, "contacts"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing basic read for calls")

	if err := testRead(ctx, conn, "calls"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing pagination for contacts")

	if err := testPagination(ctx, conn, "contacts"); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing time-based filtering/incremental sync for calls")

	if err := testCallsFiltering(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing fallback behavior for contacts filtering (expecting full read)")

	if err := testContactsFilteringIgnored(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Testing time-based filtering/incremental sync for activity")

	if err := testActivityFiltering(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *cloudtalk.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testPagination(ctx context.Context, conn *cloudtalk.Connector, objectName string) error {
	pageSize := 2

	slog.Info("Testing pagination", "object", objectName, "pageSize", pageSize)

	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
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
		return fmt.Errorf("error reading %s page 2: %w", objectName, err)
	}

	slog.Info("Page 2 results", "rows", res2.Rows)
	utils.DumpJSON(res2, os.Stdout)

	return nil
}

func testCallsFiltering(ctx context.Context, conn *cloudtalk.Connector) error {
	// Filter calls created in the last 7 days
	since := time.Now().Add(-7 * 24 * time.Hour)
	until := time.Now()

	slog.Info("Testing calls filtering (recent)", "since", since, "until", until)
	params := common.ReadParams{
		ObjectName: "calls",
		Fields:     connectors.Fields("id", "created_at"), // Request created_at to verify if returned
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading calls with range: %w", err)
	}

	slog.Info("Recent calls results", "rows", res.Rows)

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testContactsFilteringIgnored(ctx context.Context, conn *cloudtalk.Connector) error {
	// Filter contacts with a timestamp. This should be ignored by the connector
	// because we don't support client-side filtering for contacts yet,
	// and API doesn't support provider-side filtering for contacts.
	since := time.Now().Add(-1 * time.Hour)

	slog.Info("Verifying that filtering parameters are ignored for contacts", "since", since)
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id"),
		Since:      since,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading contacts with ignored parameter: %w", err)
	}

	slog.Info("Contacts read result", "rows", res.Rows)
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testActivityFiltering(ctx context.Context, conn *cloudtalk.Connector) error {
	// Filter activity created in the last 7 days.
	since := time.Now().Add(-7 * 24 * time.Hour)
	until := time.Now()

	slog.Info("Testing activity filtering (recent)", "since", since, "until", until)
	params := common.ReadParams{
		ObjectName: "activity",
		Fields:     connectors.Fields("id"),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading activity with range: %w", err)
	}

	slog.Info("Recent activity results", "rows", res.Rows)

	utils.DumpJSON(res, os.Stdout)

	return nil
}
