package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/workday"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWorkdayConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"workers"})
	if err != nil {
		utils.Fail("error listing metadata for workday", "error", err)
	}

	fmt.Println("workday metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}
