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
	"github.com/amp-labs/connectors/internal/datautils"
	gh "github.com/amp-labs/connectors/providers/greenhouse"
	connTest "github.com/amp-labs/connectors/test/greenhouse"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGreenhouseConnector(ctx)

	if err := testRead(ctx, conn, "applications",
		connectors.Fields("id", "candidate_id", "status", "updated_at"),
	); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "users",
		connectors.Fields("id", "first_name", "last_name", "site_admin", "updated_at"),
	); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(
	ctx context.Context, conn *gh.Connector, objectName string, fields datautils.StringSet,
) error {
	slog.Info("Reading " + objectName + "...")

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     fields,
	})
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	slog.Info(fmt.Sprintf("%s: rows=%d, done=%t, nextPage=%q", objectName, res.Rows, res.Done, res.NextPage))

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
