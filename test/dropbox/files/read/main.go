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

	params := common.ReadParams{
		ObjectName: "files",
		Fields:     connectors.Fields("id", "name", "server_modified", ".tag"),
		PageSize:   25,
	}

	slog.Info("=== Reading files (list_folder root) ===")
	testscenario.ReadThroughPages(ctx, conn, params)

	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading files", "error", err)
	}
	if res.Rows == 0 {
		slog.Warn("Skipped metadata-vs-read check; list_folder returned no entries")

		return
	}

	testscenario.ValidateMetadataContainsRead(ctx, conn, "files", nil)
}
