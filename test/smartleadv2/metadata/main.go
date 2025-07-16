package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/smartleadv2"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetSmartleadV2Connector(ctx)

	testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "client")
	testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "campaigns")
}
