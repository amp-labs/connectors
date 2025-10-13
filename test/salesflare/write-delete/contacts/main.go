package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/salesflare"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	FirstName string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesflareConnector(ctx)

	// Name
	firstName := gofakeit.Name()
	updatedFirstName := gofakeit.Name()
	// Last name
	lastName := gofakeit.Name()
	updatedLastName := gofakeit.Name()
	// Email
	email := gofakeit.Email()
	updatedEmail := gofakeit.Email()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName: firstName,
			Lastname:  lastName,
			Email:     email,
		},
		payload{
			FirstName: updatedFirstName,
			Lastname:  updatedLastName,
			Email:     updatedEmail,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "firstname", "lastname", "email"),
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: email,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"firstname": updatedFirstName,
				"lastname":  updatedLastName,
				"email":     updatedEmail,
			},
		},
	)
}
