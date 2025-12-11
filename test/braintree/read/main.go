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
	"github.com/amp-labs/connectors/providers/braintree"
	braintreeTest "github.com/amp-labs/connectors/test/braintree"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := braintreeTest.GetBraintreeConnector(ctx)

	if err := testRead(ctx, conn, "customers", time.Time{}, time.Time{}); err != nil {
		slog.Error(err.Error())
	}

	if err := testRead(
		ctx,
		conn,
		"transactions",
		time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
	); err != nil {
		slog.Error(err.Error())
	}
}

func testRead(ctx context.Context, conn *braintree.Connector, objectName string, since time.Time, until time.Time) error {
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(""),
		Since:      since,
		Until:      until,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", objectName, err)
	}

	// Print the results.
	utils.DumpJSON(res, os.Stdout)

	return nil
}
