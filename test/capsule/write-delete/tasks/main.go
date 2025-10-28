package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/capsule"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Party       Party  `json:"party"`
	Description string `json:"description"`
	DueOn       string `json:"dueOn"`
}

type Party struct {
	ID int `json:"id"`
}

const partyID = 277242033

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCapsuleConnector(ctx)

	description := gofakeit.Name()
	updatedDescription := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"tasks",
		Payload{
			Party:       Party{ID: partyID},
			Description: description,
			DueOn:       "2025-05-20",
		},
		Payload{
			Party:       Party{ID: partyID},
			Description: updatedDescription,
			DueOn:       "2025-05-20",
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "description"),
			SearchBy: testscenario.Property{
				Key:   "description",
				Value: description,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"description": updatedDescription,
			},
		},
	)
}
