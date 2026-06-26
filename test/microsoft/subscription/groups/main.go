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

	displayName := gofakeit.Name()
	mailNickname := "someMailNickname"

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &microsoft.SubscriptionRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"groups": {
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
				/* Create group */ {
					ObjectName: "groups",
					Method:     testscenario.ConnectorMethodCreate,
					Payload: payload{
						DisplayName:     displayName,
						MailEnabled:     false,
						MailNickname:    mailNickname,
						SecurityEnabled: true,
					},
				},
				/* Remove group */ {
					ObjectName: "groups",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "displayName"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "displayname", Value: displayName},
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
	DisplayName     string `json:"displayName"`
	MailEnabled     bool   `json:"mailEnabled"`
	MailNickname    string `json:"mailNickname"`
	SecurityEnabled bool   `json:"securityEnabled"`
}
