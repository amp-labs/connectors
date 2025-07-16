package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetGongConnector(ctx)

	testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "users")
}
