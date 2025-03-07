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
	dx "github.com/amp-labs/connectors/providers/dixa"

	"github.com/amp-labs/connectors/test/dixa"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := dixa.GetConnector(ctx)

	if err := testRead(ctx, conn, "agents", []string{"id", "createdAt", "displayName"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "business-hours/schedules", []string{"id", "name"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "queues", []string{"name", "slaCalculationMethod", "id"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *dx.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		// NextPage:   "/v1/agents?pageKey=ZnJvbT1leUpwWkNJNklqaGhNVEpqTkdFM0xURTVZVEl0TkRKak9TMDVaV00yTFRKaU5UazVOamt6WXpZNU5pSXNJbTl5WjE5cFpDSTZJamMxT1dJM056WXhMV1U1TVRJdE5Ea3hZeTFpWmpZd0xURTBNR1V6WWpFelpqSXlaaUlzSW5WelpYSmZkSGx3WlNJNkltMWxiV0psY2lKOSZwYWdlTGltaXQ9MSZkaXI9Zm9yd2FyZA%3D%3D",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objectName, err)
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
