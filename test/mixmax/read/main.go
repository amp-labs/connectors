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
	mx "github.com/amp-labs/connectors/providers/mixmax"
	"github.com/amp-labs/connectors/test/mixmax"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := mixmax.GetConnector(ctx)

	if err := testRead(ctx, conn, "snippets", []string{"_id", "createdAt", "title"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "appointmentlinks/me", []string{"id"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "integrations/commands", []string{"name", "commands", "id"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *mx.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		// NextPage:   "WyJfX21peG1heF9fdW5kZWZpbmVkX18iLHisJG9pZCI6IjU2ZGUwYjdjZWExZGU5YjY2YTUzZjRmYiJ9XQ",
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
