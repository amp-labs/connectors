package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := slack.NewConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"conversations", "users", "auth.teams"})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(m, os.Stdout)
}
