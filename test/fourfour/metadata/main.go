package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/fourfour"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := fourfour.GetFourFourConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"Labels"})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(m, os.Stdout)
}
