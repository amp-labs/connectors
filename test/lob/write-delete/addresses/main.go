package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/lob"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type CreatePayload struct {
	Name         string `json:"name"`
	AddressLine1 string `json:"address_line1"`
	AddressCity  string `json:"address_city"`
	AddressState string `json:"address_state"`
	AddressZip   string `json:"address_zip"`
	Email        string `json:"email"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetLobConnector(ctx)

	email := gofakeit.Email()
	testscenario.ValidateCreateDelete(ctx, conn,
		"addresses",
		CreatePayload{
			Name:         gofakeit.Name(),
			Email:        email,
			AddressLine1: "210 King St",
			AddressCity:  "San Francisco",
			AddressState: "CA",
			AddressZip:   "94107",
		},
		testscenario.CRDTestSuite{
			ReadFields: datautils.NewSet("name", "email"),
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: email,
			},
			RecordIdentifierKey: "id",
		})
}
