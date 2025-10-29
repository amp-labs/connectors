package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/sellsy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSellsyConnector(ctx)

	firstName := gofakeit.Name()
	updatedFirstName := gofakeit.Name()
	lastName := gofakeit.Name()
	updatedLastName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName: firstName,
			LastName:  lastName,
		},
		payload{
			FirstName: updatedFirstName,
			LastName:  updatedLastName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "first_name", "last_name"),
			SearchBy: testscenario.Property{
				Key:   "last_name",
				Value: lastName,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"first_name": updatedFirstName,
				"last_name":  updatedLastName,
			},
		},
	)
}
