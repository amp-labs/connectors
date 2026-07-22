package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	res, err := conn.ListObjectMetadata(ctx, []string{
		"contacts",
		"meetings",
		"notes",
		"tasks",
		"users",
	})
	if err != nil {
		utils.Fail("error listing metadata from Jump", "error", err)
	}

	slog.Info("Listing metadata..")
	utils.DumpJSON(res, os.Stdout)
}
