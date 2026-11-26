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
		ObjectName: "file_requests",
		Fields:     connectors.Fields("id", "title", "is_open", "url", "destination"),
	}

	slog.Info("=== Reading file_requests ===")
	testscenario.ReadThroughPages(ctx, conn, params)

	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading file_requests", "error", err)
	}
	if res.Rows == 0 {
		slog.Warn("Skipped metadata-vs-read check; no file requests returned")

		return
	}

	testscenario.ValidateMetadataContainsRead(ctx, conn, "file_requests", nil)
}
