package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/breezy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBreezyConnector(ctx)

	slog.Info("=== Reading positions ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "positions",
		Fields:     connectors.Fields("_id", "name", "state"),
	})

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "positions",
		Fields:     connectors.Fields("_id", "name", "state"),
	})
	if err != nil {
		utils.Fail("error reading positions", "error", err)
	}

	if res.Rows == 0 {
		slog.Warn("Skipped metadata-vs-read check; no published positions returned")

		return
	}

	testscenario.ValidateMetadataContainsRead(ctx, conn, "positions", nil)
}
