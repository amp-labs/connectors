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
	gr "github.com/amp-labs/connectors/providers/groove"
	"github.com/amp-labs/connectors/test/groove"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := groove.GetConnector(ctx)

	if err := testRead(ctx, conn, "tickets", []string{"id", "created_at", "number"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "widgets", []string{"id", "name"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "agents", []string{"email", "first_name", "id"}); err != nil {
		slog.Error(err.Error())
	}

}

func testRead(ctx context.Context, conn *gr.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		// NextPage:   "https://api.groovehq.com/v1/tickets?page=2",
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
