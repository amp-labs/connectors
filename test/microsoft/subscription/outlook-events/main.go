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

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	subject := gofakeit.Name()
	updatedSubject := gofakeit.Name()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: microsoft.SubscribeRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/events": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeDelete,
							},
						},
					},
				}
			},
			ExpectedWebhookCalls: 2,
			Operations: []testscenario.ConnectorOperation{
				/* Create Outlook calendar event */ {
					ObjectName: "me/events",
					Method:     testscenario.ConnectorMethodCreate,
					Payload:    payload{Subject: subject},
				},
				/* Update Outlook calendar event */ {
					ObjectName: "me/events",
					Method:     testscenario.ConnectorMethodUpdate, // We are not listening to this event.
					Payload:    payload{Subject: updatedSubject},
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "subject"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "subject", Value: subject},
					},
				},
				/* Remove Outlook calendar event */ {
					ObjectName: "me/events",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "subject"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "subject", Value: updatedSubject},
					},
				},
			},
			WebhookRouter:          subscription.NewWebhookRouter(),
			VerificationParams:     nil,
			AutoRemoveSubscription: true,
		},
	)
}

type payload struct {
	Subject string `json:"subject"`
}
