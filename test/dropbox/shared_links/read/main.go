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
		ObjectName: "shared_links",
		Fields:     connectors.Fields("id", "url", "name", ".tag"),
	}

	slog.Info("=== Reading shared_links ===")
	testscenario.ReadThroughPages(ctx, conn, params)

	res, err := conn.Read(ctx, params)
	if err != nil {
		utils.Fail("error reading shared_links", "error", err)
	}
	if res.Rows == 0 {
		slog.Warn("Skipped metadata-vs-read check; no shared links returned")

		return
	}

	testscenario.ValidateMetadataContainsRead(ctx, conn, "shared_links", nil)
}
