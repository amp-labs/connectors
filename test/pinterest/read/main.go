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
	ap "github.com/amp-labs/connectors/providers/pinterest"
	"github.com/amp-labs/connectors/test/pinterest"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := pinterest.GetConnector(ctx)

	if err := testRead(ctx, conn, "pins"); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "boards"); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "media"); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *ap.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(""),
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
