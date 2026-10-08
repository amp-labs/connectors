package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	slog.Info("=== Reading survey_folders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "survey_folders",
		Fields:     connectors.Fields("id", "title", "num_surveys"),
		PageSize:   2,
	})
}
