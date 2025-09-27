package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	cal "github.com/amp-labs/connectors/providers/calendly"
	"github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := calendly.GetConnector(ctx)

	if err := testRead(ctx, conn, "scheduled_events", []string{
		"uri", "name", "status", "start_time", "end_time", "event_type", "created_at",
	}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *cal.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objectName, err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
} 