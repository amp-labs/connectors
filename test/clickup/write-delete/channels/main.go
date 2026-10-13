package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/clickup"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetClickupConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()
	description := gofakeit.Name()
	updatedDescription := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"channels",
		Payload{
			Name:        name,
			Description: description,
		},
		Payload{
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
		})
}
