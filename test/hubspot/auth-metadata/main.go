package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	postAuthInfo, err := conn.GetPostAuthInfo(ctx)

	if err != nil {
		utils.Fail("error getting post auth info", "error", err)
	}

	utils.DumpJSON(postAuthInfo, os.Stdout)
}
