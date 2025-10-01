package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZoomConnector(ctx)
	defer utils.Close(conn)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"users", "groups"})
	if err != nil {
		utils.Fail("error listing metadata for zoom", "error", err)
	}

	fmt.Println("zoom metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}
