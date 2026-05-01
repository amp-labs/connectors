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
	bodyData := gofakeit.Name()
	from := gofakeit.Username()
	to := gofakeit.Username()
	updatedSubject := gofakeit.Name()
	updatedBodyData := gofakeit.Name()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: microsoft.SubscribeRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeUpdate,
								common.SubscriptionEventTypeDelete,
							},
						},
						//"butterflies": {
						//	Events: []common.SubscriptionEventType{
						//		common.SubscriptionEventTypeCreate,
						//	},
						//},
					},
				}
			},
			ExpectedWebhookCalls: 3,
			Operations: []testscenario.ConnectorOperation{
				/* Create outlook message */ {
					ObjectName: "me/messages",
					Method:     testscenario.ConnectorMethodCreate,
					Payload: payload{
						Subject:      subject,
						Body:         body{Content: bodyData, ContentType: TextContentType},
						From:         &recipient{EmailAddress: address{Address: from + "@test.com", Name: from}},
						ToRecipients: []recipient{{EmailAddress: address{Address: to + "@test.com", Name: to}}},
					},
				},
				/* Update outlook message */ {
					ObjectName: "me/messages",
					Method:     testscenario.ConnectorMethodUpdate,
					Payload: payload{
						Subject: updatedSubject,
						Body:    body{Content: updatedBodyData, ContentType: TextContentType},
					},
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "subject"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "subject", Value: subject},
					},
				},
				/* Remove outlook message */ {
					ObjectName: "me/messages",
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
	Subject      string      `json:"subject,omitempty"`
	Body         body        `json:"body,omitempty"`
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
