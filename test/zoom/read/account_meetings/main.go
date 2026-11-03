package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZoomConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "account_meetings",
		Fields:     connectors.Fields("email"),
	})
}
