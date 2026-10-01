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

	slog.Info("=== Reading applicant_tracking_jobs ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "applicant_tracking_jobs",
		Fields:     connectors.Fields("id", "title", "status"),
	})
}
