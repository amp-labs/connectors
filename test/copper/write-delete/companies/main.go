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

type companyPayload struct {
	Name    string `json:"name"`
	Details string `json:"details"`
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
	details := gofakeit.Name()
	updatedDetails := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"companies",
		companyPayload{
			Name:    name,
			Details: details,
		},
		companyPayload{
			Name:    updatedName,
			Details: updatedDetails,
		},
		testscenario.CRUDTestSuite{
			ReadFields:       datautils.NewSet("id", "name", "details"),
			WaitBeforeSearch: 40 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name":    updatedName,
				"details": updatedDetails,
			},
		},
	)
}
