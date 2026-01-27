package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type accountPayload struct {
	Name          string `json:"Name"`
	AccountNumber string `json:"AccountNumber"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	name := "TEST-" + gofakeit.Company()
	updatedName := "UPDATED-" + gofakeit.Company()

	accountNumber := gofakeit.Numerify("######")
	updatedAccountNumber := gofakeit.Numerify("######")

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"Account",
		accountPayload{
			Name:          name,
			AccountNumber: accountNumber,
		},
		accountPayload{
			Name:          updatedName,
			AccountNumber: updatedAccountNumber,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet(
				"id",
				"name",
				"accountnumber",
				"isdeleted",
			),
			WaitBeforeSearch: 30 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: name,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name":          updatedName,
				"accountnumber": updatedAccountNumber,
			},
		},
	)
}
