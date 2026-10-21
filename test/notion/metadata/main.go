package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/notion"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := notion.GetNotionConnector(ctx)

	result, err := conn.ListObjectMetadata(ctx, []string{
		"users",
		"file_uploads",
	})
	if err != nil {
		utils.Fail("error listing Notion metadata", "error", err)
	}

	utils.DumpJSON(result, os.Stdout)
}
