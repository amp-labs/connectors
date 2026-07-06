package testscenario

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

// WebhookRouter holds a set of conditional webhook handlers that dispatch webhook requests
// to different handlers based on routing conditions.
//
// During webhook processing, each handler is invoked in order; the first handler
// that returns true (indicating it handled the request) stops the routing process.
// If no handler returns true, the default handler is invoked.
type WebhookRouter struct {
	// Handlers is the ordered list of webhook handlers. Each handler is invoked
	// in order until one returns true (indicating the request was handled).
	Routes []WebhookRouteFunc
}

// WebhookRouteFunc is the function signature for a single webhook route handler.
//
// It receives the HTTP response writer, the original HTTP request, and a fully
// pre-read copy of the request body.
//
// IMPORTANT: requestBody is guaranteed to contain the complete request payload
// and is safe to use without re-reading or consuming request.Body. The router
// reads and buffers the body once before invoking any route handlers, ensuring
// that all handlers receive the same byte slice.
//
// Notes:
//   - Handlers must NOT read from request.Body.
type WebhookRouteFunc func(writer http.ResponseWriter, request *http.Request, requestBody []byte) bool

// RunWebhookConsumer starts a long‑running webhook consumer that listens for
// incoming change‑notification messages from a provider and prints them to stdout.
//
// The server routes incoming messages according to the provided webhookRouter:
// if a route matches the request, the associated handler is called;
// otherwise, the message is logged as JSON to os.Stdout.
//
// Verification parameters are only relevant for connectors that perform validation.
// The caller can stop this loop only by cancelling the provided context. When ctx.Done()
// is closed, the webhook server is shut down and the function returns.
func RunWebhookConsumer(
	ctx context.Context,
	conn testconn.TestableWebhookMessageVerifier,
	webhookRouter WebhookRouter,
	verificationParams *common.VerificationParams,
) {
	messageChannel := make(chan webhookMessageResult)
	_, shutdown := startWebhookHandler(ctx, conn,
		webhookRouter, verificationParams, messageChannel,
	)
	defer shutdown()

	for {
		select {
		case message := <-messageChannel:
			if message.Error == "" {
				utils.DumpJSON(message.Body, os.Stdout)
			} else {
				utils.DumpJSON(message.Error, os.Stdout)
			}
			// Infinite loop.
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping...")
			return
		}
	}
}
