package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceAccountEngagementConnector(ctx)

	email := gofakeit.Email()
	firstName := gofakeit.Name()
	lastName := gofakeit.Name()
	newEmail := gofakeit.Email()
	newFirstName := gofakeit.Name()
	newLastName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"Prospects",
		Payload{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		},
		Payload{
			Email:     newEmail,
			FirstName: newFirstName,
			LastName:  newLastName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "email", "firstName", "lastName"),
			SearchBy: testscenario.Property{
				Key:   "email", // returned fields are in lowercase
				Value: email,
			},
			RecordIdentifierKey: "id", // returned fields are in lowercase
			UpdatedFields: map[string]string{
				"email":     newEmail,
				"firstname": newFirstName,
				"lastname":  newLastName,
			},
		},
	)
}
