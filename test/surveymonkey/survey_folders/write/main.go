package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	title := fmt.Sprintf("Amp Integration Folder %s", gofakeit.UUID())

	testscenario.ValidateCreate(ctx, conn, "survey_folders",
		map[string]any{"title": title},
		testscenario.CRDTestSuite{
			ReadFields:          datautils.NewSet("id", "title"),
			RecordIdentifierKey: "id",
			SearchBy: testscenario.Property{
				Key:   "title",
				Value: title,
			},
		},
	)
}
