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

	name := fmt.Sprintf("Amp Integration %s", gofakeit.Word())

	createPayload := map[string]any{"name": name}
	updatePayload := map[string]any{"name": name + " (Updated)"}

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "contact_lists", createPayload, updatePayload,
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": name + " (Updated)",
			},
		},
	)
}
