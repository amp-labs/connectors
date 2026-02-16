package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/salesfinity"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contact-lists/csv",
		Fields:     connectors.Fields("_id"),
		PageSize:   3,
		Since:      time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC),
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "call-log",
		Fields:     connectors.Fields("_id"),
		PageSize:   100,
		Since:      time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
	})

}
