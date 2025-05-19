package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/insightly"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type CreatePayload struct {
	Salutation string `json:"SALUTATION"`
	Firstname  string `json:"FIRST_NAME"`
	Lastname   string `json:"LAST_NAME"`
}

type UploadPayload struct {
	Salutation string `json:"SALUTATION"`
	Firstname  string `json:"FIRST_NAME"`
	Lastname   string `json:"LAST_NAME"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetInsightlyConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"Leads",
		CreatePayload{
			Salutation: "Mr",
			Firstname:  "Seamus",
			Lastname:   "Ramirez",
		},
		UploadPayload{
			Salutation: "Ms",
			Firstname:  "Pamela",
			Lastname:   "Huber",
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("lead_id", "first_name", "last_name", "salutation"),
			SearchBy: testscenario.Property{
				Key:   "first_name",
				Value: "Seamus",
			},
			RecordIdentifierKey: "lead_id",
			UpdatedFields: map[string]string{
				"salutation": "Ms",
				"first_name": "Pamela",
				"last_name":  "Huber",
			},
		})
}
