package testscenario

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testscenario/internal/webhook"
)

type (
	// WebhookProcessor handles verified webhook events and optional request
	// interception in test scenarios.
	WebhookProcessor = webhook.Processor

	// WebhookInterceptorFunc handles provider-specific webhook requests before
	// they are verified. It can be used to implement validation challenges or
	// other custom request handling.
	WebhookInterceptorFunc = webhook.InterceptorFunc
)

// RunWebhookConsumer starts a long‑running webhook consumer that listens for
// incoming messages from a provider.
//
// The webhook.Server runs in the background, passes events to the connector for verification,
// and then sends them to the Processor for handling (usually printing to stdout).
//
// Verification parameters are only relevant for connectors that perform validation.
// The caller can stop this loop only by cancelling the provided context. When ctx.Done()
// is closed, the webhook server is shut down and the function returns.
func RunWebhookConsumer(ctx context.Context,
	processor *WebhookProcessor,
	conn testconn.TestableWebhookMessageVerifier,
	verificationParams *common.VerificationParams,
) {
	server := webhook.CreateServer(ctx, processor, conn, verificationParams)
	_, shutdown := server.Start(ctx)
	defer shutdown()

	processor.Run(ctx, func(event webhook.Event) bool {
		if event.Error == "" {
			utils.DumpJSON(event.Body, os.Stdout)
		} else {
			utils.DumpJSON(event.Error, os.Stdout)
		}

		return false // never stops
	})
}
