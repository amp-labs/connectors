package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/wealthbox"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWealthboxConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"contacts",
		"tasks",
		"events",
		"opportunities",
		"projects",
		"notes",
		"workflows",
		"workflow_templates",
		"comments",
		"activity",
		"users",
		"teams",
		"user_groups",
		"contact_roles",
		"custom_fields",
		"tags",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}
