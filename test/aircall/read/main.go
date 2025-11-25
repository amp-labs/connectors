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
	"github.com/amp-labs/connectors/providers/aircall"
	testAircall "github.com/amp-labs/connectors/test/aircall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testAircall.GetAircallConnector(ctx)

	if err := testRead(ctx, conn, "calls"); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(ctx, conn, "users"); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *aircall.Connector, objectName string) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	// Print the results.
	utils.DumpJSON(res, os.Stdout)

	return nil
}
