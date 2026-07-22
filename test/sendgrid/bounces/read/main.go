package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/sendgrid"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSendGridConnector(ctx)

	slog.Info("=== Reading bounces ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "bounces",
		Fields:     connectors.Fields("email", "reason", "status"),
	})
}
