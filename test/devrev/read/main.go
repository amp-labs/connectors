package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	devrevtest "github.com/amp-labs/connectors/test/devrev"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := devrevtest.GetConnector(ctx)
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("display_id"),
		PageSize:   3,
		Since:      time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "articles",
		Fields:     connectors.Fields("display_id"),
		PageSize:   3,
		Since:      time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	})
}
