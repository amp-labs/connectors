package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/netsuite"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNetsuiteRESTAPIConnector(ctx)

	objectNames := []string{
		"customer",
		"contact",
	}

	res, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail("error getting metadata from NetSuite REST API", "error", err)
	}

	slog.Info("Getting metadata for NetSuite REST API objects..")
	utils.DumpJSON(res, os.Stdout)
}
