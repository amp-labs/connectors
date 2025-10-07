package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/teamwork"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Name string `json:"name"`
	City string `json:"city"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetTeamworkConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()
	city := gofakeit.City()
	updatedCity := gofakeit.City()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"companies",
		Payload{
			Name: name,
			City: city,
		},
		Payload{
			Name: updatedName,
			City: updatedCity,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "name", "city"),
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
				"city": updatedCity,
			},
		},
	)
}
