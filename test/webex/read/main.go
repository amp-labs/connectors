package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/webex"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWebexConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "people",
		Fields:     connectors.Fields("id", "displayName"),
		PageSize:   10,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "groups",
		Fields:     connectors.Fields("id", "displayName"),
		PageSize:   1,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "roles",
		Fields:     connectors.Fields("id", "displayName"),
		PageSize:   1,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "events",
		Fields:     connectors.Fields("id", "resource", "start"),
		PageSize:   10,
		Since:      time.Date(2025, 12, 5, 21, 37, 4, 126, time.UTC),
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "organizations",
		Fields:     connectors.Fields("id", "displayName", "created"),
		PageSize:   10,
	})
}
