package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/dynamicsbusiness"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Type        string `json:"type"`
	DisplayName string `json:"displayName"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetDynamicsBusinessCentralConnector(ctx)

	displayName := gofakeit.Name()
	newDisplayName := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"Contacts",
		Payload{
			Type:        "Person",
			DisplayName: displayName,
		},
		Payload{
			Type:        "Person",
			DisplayName: newDisplayName,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "type", "displayName"),
			SearchBy: testscenario.Property{
				Key:   "displayname", // returned fields are in lowercase
				Value: displayName,
			},
			RecordIdentifierKey: "id", // returned fields are in lowercase
			UpdatedFields: map[string]string{
				"displayname": newDisplayName,
			},
		},
	)
}
