package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot"
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

	propMsg := hubspot.WebhookMessage{}

	if err := json.Unmarshal([]byte(samplePropertyChange), &propMsg); err != nil {
		utils.Fail("error unmarshalling property change message", "error", err)
	}

	recordId, err := strconv.Atoi(writeResult.RecordId)
	if err != nil {
		utils.Fail("error converting record id to int", "error", err)
	}

	propMsg.ObjectId = recordId

	recordResult, err := conn.GetRecordFromWebhookMessage(ctx, &propMsg)
	if err != nil {
		utils.Fail("error getting record from webhook message", "error", err)
	}

	utils.DumpJSON(recordResult, os.Stdout)
}
