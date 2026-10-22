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

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "posts",
		Fields: connectors.Fields("id", "date_gmt", "comment_status",
			"$['title']['rendered']",
		),
		// Since:      utils.Timestamp("2026-10-19T19:31:45Z"),
		PageSize: 2,
	})
}
