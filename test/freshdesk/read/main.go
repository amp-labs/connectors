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
	fd "github.com/amp-labs/connectors/providers/freshdesk"
	"github.com/amp-labs/connectors/test/freshdesk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := freshdesk.GetFreshdeskConnector(ctx)

	if err := testRead(ctx, conn, "tickets", []string{"id", "name"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "contacts", []string{"name", "id", "mobile", "job_title"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "companies", []string{"id", "name", "note"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *fd.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      time.Now().Add(-50 * time.Hour),
		NextPage:   "",
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
