package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetConnectWiseConnector(ctx)

	testscenario.SubscriptionCreateUpdateDelete(ctx, conn,
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: &connectwise.SubscriptionRequest{
					WebhookURL: webhookURL,
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"projects": {},
					"invoices": {},
				},
			}
		},
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: &connectwise.SubscriptionRequest{
					WebhookURL: webhookURL,
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"invoices":        {},
					"project/tickets": {},
					"contacts":        {},
				},
			}
		},
	)
}
