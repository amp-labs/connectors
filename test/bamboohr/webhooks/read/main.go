package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/bamboohr"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBambooHRConnector(ctx)

	slog.Info("=== Reading webhooks ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "webhooks",
		Fields:     connectors.Fields("id", "name", "url", "created", "lastSent"),
	})
}
