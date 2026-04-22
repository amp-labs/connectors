package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/procoresandbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn, err := procoresandbox.NewConnector(ctx)
	if err != nil {
		utils.Fail("error creating procore connector", "error", err)
	}

	m, err := conn.ListObjectMetadata(ctx, []string{"comm-handling/states", "contacts", "meetings"})
	if err != nil {
		utils.Fail("error listing metadata for Procore", "error", err)
	}

	// Print the results
	utils.DumpJSON(m.Result, os.Stdout)
	fmt.Println("Errors: ", m.Errors)
}
