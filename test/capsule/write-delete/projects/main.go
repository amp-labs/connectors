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
	Name        string `json:"name"`
	Description string `json:"description"`
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

	name := gofakeit.Name()
	description := gofakeit.Name()
	updatedName := gofakeit.Name()
	updatedDescription := gofakeit.Name()

	// Object names can be used interchangeably.
	// https://developer.capsulecrm.com/v2/operations/Case
	objectAliases := []string{"projects", "kases"}

	for _, objectName := range objectAliases {
		testscenario.ValidateCreateUpdateDelete(ctx, conn,
			objectName,
			Payload{
				Party:       Party{ID: partyID},
				Name:        name,
				Description: description,
			},
			Payload{
				Party:       Party{ID: partyID},
				Name:        updatedName,
				Description: updatedDescription,
			},
			testscenario.CRUDTestSuite{
				ReadFields: datautils.NewSet("id", "name", "description"),
				SearchBy: testscenario.Property{
					Key:   "name",
					Value: name,
				},
				RecordIdentifierKey: "id",
				UpdatedFields: map[string]string{
					"name":        updatedName,
					"description": updatedDescription,
				},
			},
		)
	}
}
