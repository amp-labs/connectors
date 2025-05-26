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
	pd "github.com/amp-labs/connectors/providers/podium"
	"github.com/amp-labs/connectors/test/podium"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := podium.GetConnector(ctx)

	if err := testRead(ctx, conn, "contacts", []string{"id", "updatedAt", "phoneNumbers"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "webhooks", []string{"updatedAt", "url"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "campaigns", []string{"name", "message", "uid"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *pd.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      time.Now().Add(-10000 * time.Hour),
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
