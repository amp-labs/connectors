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
	ap "github.com/amp-labs/connectors/providers/avoma"
	"github.com/amp-labs/connectors/test/avoma"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := avoma.GetAvomaConnector(ctx)

	if err := testRead(ctx, conn, "users", time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(
		ctx,
		conn,
		"meetings",
		time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.June, 30, 0, 0, 0, 0, time.UTC),
	); err != nil {
		slog.Error(err.Error())
	}

}

func testRead(ctx context.Context, conn *ap.Connector, objectName string, since time.Time, until time.Time) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(""),
		Since:      since,
		Until:      until,
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
