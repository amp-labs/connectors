package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/zendesksupport"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetZendeskSupportConnector(ctx)

	testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "brands")
}
