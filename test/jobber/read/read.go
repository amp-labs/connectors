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
	ap "github.com/amp-labs/connectors/providers/jobber"
	"github.com/amp-labs/connectors/test/jobber"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := jobber.GetJobberConnector(ctx)

	if err := testRead(ctx, conn, "apps", []string{"id", "name"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "clients", []string{"id", "name", "updatedAt"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "visits", []string{"id", "title"}); err != nil {
		slog.Error(err.Error())
	}

	// Incremental reads. Clients filter natively on updatedAt, visits on
	// createdAt (no updatedAt exists), and jobs sort by UPDATED_AT descending
	// with a client-side cutoff.
	//
	// Note: Jobber throttles on query cost (10k point bucket, 500/s restore)
	// and the full jobs query alone costs ~9.6k points, so back-to-back reads
	// may return "Throttled".
	since := time.Now().AddDate(0, 0, -7)

	if err := testIncrementalRead(ctx, conn, "clients", []string{"id", "name", "updatedAt"}, since); err != nil {
		slog.Error(err.Error())
	}

	if err := testIncrementalRead(ctx, conn, "jobs", []string{"id", "title", "updatedAt"}, since); err != nil {
		slog.Error(err.Error())
	}

	if err := testIncrementalRead(ctx, conn, "visits", []string{"id", "title", "createdAt"}, since); err != nil {
		slog.Error(err.Error())
	}
}

func testIncrementalRead(
	ctx context.Context, conn *ap.Connector, objectName string, fields []string, since time.Time,
) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      since,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s incrementally: %w", objectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	if _, err := os.Stdout.Write(jsonStr); err != nil {
		return fmt.Errorf("error writing JSON: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}

	return nil
}

func testRead(ctx context.Context, conn *ap.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	if _, err := os.Stdout.Write(jsonStr); err != nil {
		return fmt.Errorf("error writing JSON: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}

	return nil
}
