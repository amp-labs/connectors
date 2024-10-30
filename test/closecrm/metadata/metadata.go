package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/closecrm"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := closecrm.GetCloseConnector(ctx)

	defer utils.Close(conn)

	m, err := conn.ListObjectMetadata(ctx, []string{"lead", "contact", "activity", "task"})
	if err != nil {
		utils.Fail("error listing metadata for Close", "error", err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
