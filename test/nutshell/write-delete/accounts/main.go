package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/nutshell"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type createPayload struct {
	Name string `json:"name"`
}

type updatePayload struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNutshellConnector(ctx)

	name := gofakeit.Name()
	updatedName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"accounts",
		createPayload{
			Name: name,
		},
		updatePayload{
			Op:    "replace",
			Path:  "/accounts/0/name",
			Value: updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:       datautils.NewSet("id", "name"),
			WaitBeforeSearch: 5 * time.Second,
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
