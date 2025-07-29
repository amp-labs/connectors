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

	conn := connTest.GetNetsuiteSuiteQLConnector(ctx)

	objectNames := []string{
		"transaction",
		"vendor",
	}

	res, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail("error getting metadata from NetSuite SuiteQL", "error", err)
	}

	slog.Info("Getting metadata for NetSuite SuiteQL objects..")
	utils.DumpJSON(res, os.Stdout)
}
