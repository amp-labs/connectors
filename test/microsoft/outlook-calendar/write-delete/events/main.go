package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	Subject string `json:"subject"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	subject := gofakeit.Name()
	updatedSubject := gofakeit.Name()

	// https://learn.microsoft.com/en-us/graph/api/resources/event?view=graph-rest-1.0
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"me/events",
		payload{
			Subject: subject,
		},
		payload{
			Subject: updatedSubject,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "subject"),
			SearchBy: testscenario.Property{
				Key:   "subject",
				Value: subject,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"subject": updatedSubject,
			},
		},
	)
}
