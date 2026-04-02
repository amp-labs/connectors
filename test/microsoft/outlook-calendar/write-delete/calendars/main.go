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
	Name string `json:"name"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()

	// https://learn.microsoft.com/en-us/graph/api/resources/calendar?view=graph-rest-1.0
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"calendars",
		payload{
			Name: name,
		},
		payload{
			Name: updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "name"),
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
