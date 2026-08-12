// Stage 2 live test (see docs/subscribe-onboarding/live-tests.md): subscribes to
// customer events, creates and deletes a customer through the connector, and waits
// for Stripe to deliver the resulting webhook events.
//
// Run `ngrok http 4550` first; the script prompts for the public URL.
//
// Note on verification: Stripe generates the endpoint signing secret at subscribe
// time, so it cannot be supplied to the suite up front — incoming events are printed
// with a verification error. Use the Stage 1 consumer (test/stripe/webhook) with a
// pre-created endpoint's secret to exercise VerifyWebhookMessage end-to-end.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	name := gofakeit.Name()
	email := gofakeit.Email()

	testscenario.ValidateSubscribeReceiveEvents(ctx, conn,
		testscenario.SubscribeReceiveEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: &stripe.SubscriptionRequest{
						WebhookEndPoint: webhookURL,
					},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						// Object names map onto Stripe event prefixes (customer.created, ...).
						// https://docs.stripe.com/api/events/types
						"customer": {
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
				/* Create customer */ {
					ObjectName: "customers",
					Method:     testscenario.ConnectorMethodCreate,
					Payload: map[string]any{
						"name":  name,
						"email": email,
					},
				},
				/* Remove customer */ {
					ObjectName: "customers",
					Method:     testscenario.ConnectorMethodDelete,
					SearchProcedure: testscenario.SearchProcedure{
						ReadFields:          datautils.NewSet("id", "email"),
						RecordIdentifierKey: "id",
						SearchBy:            testscenario.Property{Key: "email", Value: email},
					},
				},
			},
			WebhookProcessor:       &testscenario.WebhookProcessor{},
			VerificationParams:     nil,
			AutoRemoveSubscription: true,
		},
	)
}
