package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zoho.GetZohoConnector(ctx, providers.ModuleZohoMail)
	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail("error getting post-auth info for Zoho", "error", err)
	}

	m, err := conn.ListObjectMetadata(ctx, []string{
		"accounts",
		"messages",
	})
	if err != nil {
		utils.Fail("error listing metadata for Zoho", "error", err)
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
	fmt.Println("Errors: ", m.Errors)
}
