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
	"activities",
	"activity_descriptions",
	"lauk_civil_controlled_rates",
	"text_snippets",
	"trust_line_items",
	"users",
	"webhooks",
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := clioTest.GetClioManageConnector(ctx)

	utils.DumpJSON(objectNames, os.Stdout)

	metadata, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail("error listing metadata for Clio Manage", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}
