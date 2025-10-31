package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	if err := testRead(ctx, conn, "apps", nil); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "clients", nil); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "visits", []string{"id", "title"}); err != nil {
		slog.Error(err.Error())
	}
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
