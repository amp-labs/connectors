package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/dropbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := dropbox.GetDropboxConnector(ctx)

	result, err := conn.ListObjectMetadata(ctx, []string{
		"files",
		"file_requests",
		"shared_links",
		"shared_folders",
	})
	if err != nil {
		utils.Fail("error listing Dropbox metadata", "error", err)
	}

	utils.DumpJSON(result, os.Stdout)
}
