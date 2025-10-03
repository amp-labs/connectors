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
	dr "github.com/amp-labs/connectors/providers/drift"
	"github.com/amp-labs/connectors/test/drift"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := drift.GetConnector(ctx)

	if err := testRead(ctx, conn, "users", []string{"id", "name", "email"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "conversations/stats", []string{"conversationCount"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "scim/Users", []string{"Resources", "schemas", "id"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *dr.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
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
