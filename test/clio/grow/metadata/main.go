package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	clioTest "github.com/amp-labs/connectors/test/clio"
	"github.com/amp-labs/connectors/test/utils"
)

var objectNames = []string{ // nolint:gochecknoglobals
	"contacts",
	"custom_actions",
	"inbox_leads",
	"matters",
	"users",
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := clioTest.GetClioGrowConnector(ctx)

	utils.DumpJSON(objectNames, os.Stdout)

	metadata, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail("error listing metadata for Clio Grow", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}
