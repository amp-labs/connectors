package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/aws"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAWSConnector(ctx, providers.ModuleAWSIdentityCenter)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"Users", "Applications",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}
