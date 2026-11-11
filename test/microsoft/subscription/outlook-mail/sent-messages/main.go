package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/microsoft/subscription"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

// IMPORTANT:
// The script waits until it receives the expected number of webhook calls.
//
// To trigger the events, manually email from current mailbox.
// When the email is sent and becomes part of the outbox/sentItems,
// the script receives the corresponding webhook event and prints its details to the console.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	subject := gofakeit.Name()
	bodyData := gofakeit.Name()
	from := gofakeit.Username()
	to := gofakeit.Username()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &microsoft.SubscriptionRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"AMPERSAND-sentMessages": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
							},
						},
					},
				}
			},
			ExpectedWebhookCalls: 2,
			// The operations that trigger webhook events must be performed manually.
			// Email using current mailbox; once it sends, the corresponding
			// event is received and printed to the console.
			//
			// NOTE: the draft is created then deleted and it must not trigger the events.
			Operations: []testscenario.ConnectorOperation{
				/* Create outlook message */ {
					ObjectName: "AMPERSAND-drafts",
					Method:     testscenario.ConnectorMethodCreate,
					Payload: payload{
						Subject:      subject,
						Body:         body{Content: bodyData, ContentType: TextContentType},
						From:         &recipient{EmailAddress: address{Address: from + "@test.com", Name: from}},
						ToRecipients: []recipient{{EmailAddress: address{Address: to + "@test.com", Name: to}}},
					},
				},
				/* Remove outlook message */ {
					ObjectName: "AMPERSAND-drafts",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "subject"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "subject", Value: subject},
					},
				},
			},
			WebhookProcessor:       subscription.NewWebhookProcessor(),
			AutoRemoveSubscription: true,
		},
	)
}

type payload struct {
	Subject      string      `json:"subject,omitempty"`
	Body         body        `json:"body"`
	From         *recipient  `json:"from,omitempty"`
	ToRecipients []recipient `json:"toRecipients,omitempty"`
}

type body struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type recipient struct {
	EmailAddress address `json:"emailAddress"`
}

type address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

const TextContentType = "text"
