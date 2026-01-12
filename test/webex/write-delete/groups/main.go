package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/webex"
	"github.com/brianvoe/gofakeit/v6"
)

type groupPayload struct {
	DisplayName string `json:"displayName,omitempty"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetWebexConnector(ctx)

	displayName := gofakeit.Company() + " Group"
	updatedDisplayName := gofakeit.Company() + " Group Updated"

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"groups",
		groupPayload{
			DisplayName: displayName,
		},
		groupPayload{
			DisplayName: updatedDisplayName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "displayName"),
			RecordIdentifierKey: "id",
			WaitBeforeSearch:    2 * time.Second,
			UpdatedFields: map[string]string{
				"displayname": updatedDisplayName,
			},
		},
	)
}
