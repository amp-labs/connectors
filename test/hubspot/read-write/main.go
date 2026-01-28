package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

// Contact is a basic Hubspot contact.
type Contact struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Company   string `json:"company"`
	Website   string `json:"website"`
	Lastname  string `json:"lastname"`
	Firstname string `json:"firstname"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	email := gofakeit.Email()
	updatedPhone := gofakeit.Phone()
	updatedCompany := gofakeit.Company()
	updatedWebsite := gofakeit.URL()
	updatedFirstName := gofakeit.FirstName()
	updatedLastName := gofakeit.LastName()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		Contact{
			Email:     email,
			Phone:     gofakeit.Phone(),
			Company:   updatedCompany,
			Website:   gofakeit.URL(),
			Lastname:  gofakeit.LastName(),
			Firstname: gofakeit.FirstName(),
		},
		Contact{
			Email:     email,
			Phone:     updatedPhone,
			Company:   updatedCompany,
			Website:   updatedWebsite,
			Lastname:  updatedLastName,
			Firstname: updatedFirstName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id",
				"email", "phone", "company", "website", "lastname", "firstname"),
			WaitBeforeSearch: 20 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: email,
				Since: time.Now().Add(-5 * time.Minute),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"phone":     updatedPhone,
				"company":   updatedCompany,
				"website":   updatedWebsite,
				"lastname":  updatedLastName,
				"firstname": updatedFirstName,
			},
		},
	)
}
