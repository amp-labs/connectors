package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	pd "github.com/amp-labs/connectors/providers/procore"
	"github.com/amp-labs/connectors/test/procoresandbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn, err := procoresandbox.NewConnector(ctx)
	if err != nil {
		slog.Error("Failed to create connector", slog.Any("error", err))
		return
	}

	if err := testRead(ctx, conn, "companies", []string{"id", "is_active", "name"}, 1000); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "submittal_statuses", []string{"id", "name", "status"}, 4); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "people", []string{"id", "first_name", "user_id"}, 1000); err != nil {
		slog.Error(err.Error())
	}

}

func testRead(ctx context.Context, conn *pd.Connector, objectName string, fields []string, pageSize int) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      time.Now().Add(-10000 * time.Hour),
		PageSize:   pageSize,
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

	log.Printf("Successfully read %d records for object %s\n", len(res.Data), objectName)

	return nil
}
