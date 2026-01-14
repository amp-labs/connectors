package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/nutshell"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	Links payloadLinks `json:"links"`
	Body  string       `json:"body"`
}

type payloadLinks struct {
	ParentID string `json:"parent"`
}

// This value may change based on the data in your account.
const accountID = "15-accounts"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNutshellConnector(ctx)

	body := gofakeit.Name()

	testscenario.ValidateCreateDelete(ctx, conn,
		"notes",
		payload{
			Links: payloadLinks{
				ParentID: accountID,
			},
			Body: body,
		},
		testscenario.CRDTestSuite{
			ReadFields: datautils.NewSet("id", "body"),
			SearchBy: testscenario.Property{
				Key:   "body",
				Value: body,
			},
			RecordIdentifierKey: "id",
		},
	)
}
