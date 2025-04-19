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

type LeadCreatePayload struct {
	Salutation string `json:"SALUTATION"`
	Firstname  string `json:"FIRST_NAME"`
	Lastname   string `json:"LAST_NAME"`
}

type LeadUploadPayload struct {
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
		LeadCreatePayload{
			Salutation: "Mr",
			Firstname:  "Seamus",
			Lastname:   "Ramirez",
		},
		LeadUploadPayload{
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
