// Stage 3 live test (see docs/subscribe-onboarding/live-tests.md): exercises the
// subscription lifecycle — Subscribe, UpdateSubscription, DeleteSubscription — against
// a real Stripe account, printing the SubscriptionResult after each step.
//
// Run `ngrok http 4550` first; the script prompts for the public URL.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	createParams := func(webhookURL string) *common.SubscribeParams {
		return &common.SubscribeParams{
			Request: &stripe.SubscriptionRequest{
				WebhookEndPoint: webhookURL,
			},
			SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
				},
				"balance": {
					PassThroughEvents: []string{"balance.available"},
				},
				"billing_portal": {
					PassThroughEvents: []string{"billing_portal.configuration.created"},
				},
				"charge": {
					PassThroughEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
				},
			},
		}
	}

	// The desired state replaces the previous one: balance and billing_portal are removed.
	updateParams := func(webhookURL string) *common.SubscribeParams {
		return &common.SubscribeParams{
			Request: &stripe.SubscriptionRequest{
				WebhookEndPoint: webhookURL,
			},
			SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
				"account": {
					Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
					PassThroughEvents: []string{"account.application.authorized"},
				},
				"charge": {
					PassThroughEvents: []string{"charge.succeeded"},
				},
			},
		}
	}

	testscenario.SubscriptionCreateUpdateDelete(ctx, conn, createParams, updateParams)
}
