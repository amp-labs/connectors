package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesloftConnector(ctx)

	testscenario.ValidateMetadataContainsRead(ctx, conn, "groups", nil)
}
