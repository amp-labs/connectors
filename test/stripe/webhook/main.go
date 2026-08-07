// Stage 1 live test (see docs/subscribe-onboarding/live-tests.md): starts a local
// webhook consumer that verifies incoming Stripe webhook messages.
//
// Usage:
//  1. Expose the local consumer publicly: `ngrok http 4550`.
//  2. Create a webhook endpoint in the Stripe dashboard (or via the Stage 2/3 harnesses)
//     pointing at the public URL, and copy its signing secret (whsec_...).
//  3. Add the signing secret to the stripe creds JSON file as 'webhookSecret'
//     (or set STRIPE_WEBHOOK_SECRET).
//  4. Run this program and trigger events in Stripe (e.g. create a customer).
//     Verified event payloads are printed to the terminal.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
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

	webhookSecretField := credscanning.Field{
		Name:      "webhookSecret",
		PathJSON:  "webhookSecret",
		SuffixENV: "WEBHOOK_SECRET",
	}
	filePath := credscanning.LoadPath(providers.Stripe)
	reader := utils.MustCreateProvCredJSON(filePath, false, webhookSecretField)

	webhookSecret := reader.Get(webhookSecretField)
	if webhookSecret == "" {
		utils.Fail("webhook secret is required; add 'webhookSecret' to the stripe credentials JSON file")
	}

	verificationParams := &common.VerificationParams{
		Param: &stripe.VerificationParams{
			Secret: webhookSecret,
		},
	}

	testscenario.RunWebhookConsumer(ctx, &testscenario.WebhookProcessor{}, conn, verificationParams)
}
