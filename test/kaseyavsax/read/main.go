package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	kaseya "github.com/amp-labs/connectors/providers/kaseyavsax"
	"github.com/amp-labs/connectors/test/kaseyavsax"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := kaseyavsax.NewConnector(ctx)

	if err := testRead(ctx, conn, "devices", []string{"Identifier", "Name", "GroupId"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "assets", []string{"ComputerId", "PublicIpAddress", "SiteId"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "groups", []string{"Id", "Name", "Notes"}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "automation/tasks", []string{"Id", "Name", "Description"}); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *kaseya.Connector, objectName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		Since:      time.Now().Add(-480 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objectName, err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
