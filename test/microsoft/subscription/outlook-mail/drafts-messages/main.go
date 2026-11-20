package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/microsoft"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/microsoft/subscription"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &microsoft.SubscriptionRequest{
						WebhookURL: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"AMPERSAND-drafts": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
								common.SubscriptionEventTypeUpdate,
								common.SubscriptionEventTypeDelete,
							},
						},
					},
				}
			},
			ExpectedWebhookCalls: 3,
			// Manually Create/Update/Delete drafts.
			// Before doing that try Sending/Receiving the email.
			// Those shouldn't be printed. Then we conclude that subscribe works for drafts.
			Operations:             nil,
			WebhookProcessor:       subscription.NewWebhookProcessor(),
			AutoRemoveSubscription: true,
		},
	)
}
