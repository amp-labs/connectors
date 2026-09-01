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

	email := fmt.Sprintf("amp-integration-%s@example.com", gofakeit.UUID())

	createPayload := map[string]any{
		"first_name": gofakeit.FirstName(),
		"last_name":  gofakeit.LastName(),
		"email":      email,
	}

	updatePayload := map[string]any{
		"first_name": createPayload["first_name"],
		"last_name":  "Updated",
		"email":      email,
	}

	testscenario.ValidateCreateUpdateDelete(ctx, conn, "contacts", createPayload, updatePayload,
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "email", "first_name", "last_name"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"last_name": "Updated",
			},
		},
	)
}
