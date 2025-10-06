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

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"projects",
		Payload{
			Party:       Party{ID: 254633973},
			Name:        name,
			Description: description,
		},
		Payload{
			Party:       Party{ID: 254633973},
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
