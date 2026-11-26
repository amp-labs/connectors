package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/dropbox"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetDropboxConnector(ctx)

	tests := []common.ReadParams{
		{
			ObjectName: "files",
			Fields:     connectors.Fields("id", "name", "server_modified"),
			PageSize:   25,
		},
		{
			ObjectName: "file_requests",
			Fields:     connectors.Fields("id", "title", "is_open"),
		},
		{
			ObjectName: "shared_links",
			Fields:     connectors.Fields("id", "url", "name"),
		},
		{
			ObjectName: "shared_folders",
			Fields:     connectors.Fields("shared_folder_id", "name"),
		},
		{
			ObjectName: "mountable_shared_folders",
			Fields:     connectors.Fields("shared_folder_id", "name"),
		},
		{
			ObjectName: "received_files",
			Fields:     connectors.Fields("id", "name"),
		},
	}

	for _, tt := range tests {
		slog.Info("=== Reading " + tt.ObjectName + " ===")
		testscenario.ReadThroughPages(ctx, conn, tt)
	}
}
