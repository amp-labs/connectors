package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetOutreachConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"snippets", "duties", "phoneNumbers", "devices", "emailAddresses"})
	if err != nil {
		utils.Fail("error listing metadata for Outreach", "error", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
