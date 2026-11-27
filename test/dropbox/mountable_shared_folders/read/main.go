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
		ObjectName: "mountable_shared_folders",
		Fields:     connectors.Fields("shared_folder_id", "name"),
	}

	slog.Info("=== Reading mountable_shared_folders ===")
	testscenario.ReadThroughPages(ctx, conn, params)

	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading mountable_shared_folders", "error", err)
	}
	if res.Rows == 0 {
		slog.Warn("Skipped metadata-vs-read check; no mountable shared folders returned")

		return
	}

	testscenario.ValidateMetadataContainsRead(ctx, conn, "mountable_shared_folders", nil)
}
