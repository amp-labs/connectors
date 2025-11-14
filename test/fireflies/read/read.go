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
	ap "github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/fireflies"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := fireflies.GetFirefliesConnector(ctx)

	if err := testRead(ctx, conn, "users", []string{"user_id"}, time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "transcripts", []string{"id", "title"}, time.Date(2025, 11, 1, 0, 0, 0, 0, time.Local), time.Date(2025, 11, 14, 0, 0, 0, 0, time.Local)); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "bites", []string{"id", "name"}, time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "activeMeetings", []string{"id", "title"}, time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "userGroups", []string{""}, time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "analytics", []string{""}, time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *ap.Connector, objectName string, fields []string, since, until time.Time) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	// Print the results.
	utils.DumpJSON(res, os.Stdout)

	return nil
}
