// Live subscription lifecycle (Stage 3): Subscribe, UpdateSubscription, DeleteSubscription.
//
// Each step prints the stored result, the objects and events it covers, and the status.
// Reply does not call the webhook URL when a subscription is created, so any https URL works;
// the scenario asks for one, or reads the WEBHOOK_URL environment variable.
//
//	WEBHOOK_URL=https://example.com/hooks/reply go run ./test/reply/subscribe
//
// Everything it creates is deleted at the end. Run it only against a sandbox.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/reply"
	connTest "github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetReplyConnector(ctx)

	testscenario.SubscriptionCreateUpdateDelete(
		ctx, conn,
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: &reply.SubscriptionRequest{WebhookURL: webhookURL},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts": {
						PassThroughEvents: []string{"contact_opted_out", "email_replied"},
					},
					"sequences": {
						PassThroughEvents: []string{"autopilot_stopped"},
					},
				},
			}
		},
		// Drops email_replied, adds contact_finished and a LinkedIn alert.
		func(webhookURL string) *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: &reply.SubscriptionRequest{WebhookURL: webhookURL},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts": {
						PassThroughEvents: []string{"contact_opted_out", "contact_finished"},
					},
					"sequences": {
						PassThroughEvents: []string{"autopilot_stopped"},
					},
					"linkedin-accounts": {
						PassThroughEvents: []string{"linkedin_account_alerts"},
					},
				},
			}
		},
	)
}
