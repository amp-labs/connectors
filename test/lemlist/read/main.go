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
	ll "github.com/amp-labs/connectors/providers/lemlist"
	"github.com/amp-labs/connectors/test/lemlist"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := lemlist.GetLemlistConnector(ctx)

	if err := testRead(ctx, conn, "schedules", []string{"id", "name", "start"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "campaigns", []string{"id", "name"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "hooks", []string{"targetUrl", "campaignId", "_id"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *ll.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		// NextPage:   "3",
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
