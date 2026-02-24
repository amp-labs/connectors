package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	slog.Info("=== Reading all dial sessions ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "dialsession",
		Fields:     connectors.Fields("dialsession_id", "start_when", "call_count"),
	})
}
