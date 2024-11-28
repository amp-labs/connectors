package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
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

	reader := connTest.CredsReader()
	token := reader.Get(credscanning.Fields.AccessToken)

	postAuthInfo, err := conn.GetPostAuthInfo(ctx, &common.PostAuthInfoParams{
		AccessToken: token,
	})

	if err != nil {
		utils.Fail("error getting post auth info", "error", err)
	}

	utils.DumpJSON(postAuthInfo, os.Stdout)
}
