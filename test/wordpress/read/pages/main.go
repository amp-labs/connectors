package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/wordpress"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWordpressConnector(ctx)

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "pages",
		Fields:     connectors.Fields("$['content']['rendered']"),
		Since:      utils.Timestamp("2026-10-18T19:31:45Z"),
		PageSize:   10,
	})
}
