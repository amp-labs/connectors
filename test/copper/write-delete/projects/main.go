package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/copper"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type projectPayload struct {
	Name string `json:"name"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCopperConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"projects",
		projectPayload{
			Name: name,
		},
		projectPayload{
			Name: updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:       datautils.NewSet("id", "name"),
			WaitBeforeSearch: 40 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}
