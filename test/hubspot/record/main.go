package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

const samplePropertyChange = `{
  "appId": 4210286,
  "eventId": 100,
  "subscriptionId": 2902227,
  "portalId": 44237313,
  "occurredAt": 1730750483646,
  "subscriptionType": "contact.propertyChange",
  "attemptNumber": 0,
  "objectId": 74999542704,
  "changeSource": "CRM",
  "propertyName": "message",
  "propertyValue": "sample-value"
}`

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	// Write an artificial contact to Hubspot.
	writeResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "",
		RecordData: map[string]any{
			"email":     gofakeit.Email(),
			"phone":     gofakeit.Phone(),
			"company":   gofakeit.Company(),
			"website":   gofakeit.URL(),
			"lastname":  gofakeit.LastName(),
			"firstname": gofakeit.FirstName(),
		},
	})
	if err != nil {
		utils.Fail("error writing to hubspot", "error", err)
	}

	records, err := conn.GetRecordsByIds(ctx, common.ReadByIdsParams{
		ObjectName: "contact",
		RecordIds:  []string{writeResult.RecordId},
	})
	if err != nil {
		utils.Fail("error getting records by ids", "error", err)
	}

	utils.DumpJSON(records, os.Stdout)
}
