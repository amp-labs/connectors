package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	zi "github.com/amp-labs/connectors/providers/zoominfo"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoominfo"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetZoomInfoConnector(ctx)

	now := time.Now().UTC()
	last30d := now.Add(-30 * 24 * time.Hour)

	// Reference-data lookup: no criteria, single page.
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "industries",
		Fields:     connectors.Fields("name"),
	}); err != nil {
		slog.Error(err.Error())
	}

	// Search object, unfiltered: Since defaults to epoch so the required date
	// criterion is present and the read returns all records.
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("firstName", "lastUpdatedDate"),
		PageSize:   3,
	}); err != nil {
		slog.Error(err.Error())
	}

	// contacts updated in the last 30 days (lastUpdatedDateAfter).
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("firstName", "lastUpdatedDate"),
		Since:      last30d,
		PageSize:   3,
	}); err != nil {
		slog.Error(err.Error())
	}

	// news published in the last 30 days (pageDateMin/Max).
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "news",
		Fields:     connectors.Fields("title", "pageDate"),
		Since:      last30d,
		Until:      now,
		PageSize:   3,
	}); err != nil {
		slog.Error(err.Error())
	}

	// scoops published in the last 30 days (publishedStartDate).
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "scoops",
		Fields:     connectors.Fields("description", "publishedDate"),
		Since:      last30d,
		PageSize:   3,
	}); err != nil {
		slog.Error(err.Error())
	}

	// GET list object (may be empty/entitlement-gated on this account).
	if err := testRead(ctx, conn, common.ReadParams{
		ObjectName: "audiences",
		Fields:     connectors.Fields("name"),
		PageSize:   3,
	}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *zi.Connector, params common.ReadParams) error {
	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", params.ObjectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	if _, err := os.Stdout.Write(jsonStr); err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}

	return nil
}
