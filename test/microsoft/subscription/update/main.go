package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/microsoft"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	testscenario.SubscriptionCreateUpdateDelete(ctx, conn,
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: microsoft.SubscribeRequest{
					WebhookURL: webhookURL,
				},
				RegistrationResult: nil, // TODO
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
					"me/events": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
				},
			}
		},
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: microsoft.SubscribeRequest{
					WebhookURL: webhookURL,
				},
				RegistrationResult: nil, // TODO
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"me/messages": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeUpdate,
							common.SubscriptionEventTypeDelete,
						},
					},
					"me/events": {
						Events: []common.SubscriptionEventType{ // no change
							common.SubscriptionEventTypeCreate,
						},
					},
				},
			}
		},
	)
}
