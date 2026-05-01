package testscenario

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
)

// WebhookRouter holds a set of conditional HTTP routes that dispatch webhook requests
// to different handlers based on the RoutingCondition.
//
// During webhook processing, each Route is tested in order; the first handler
// whose condition returns true is called. If no condition matches, the default
// behavior is to log the message as a JSON object to stdout.
type WebhookRouter struct {
	Routes []Route
}

// RoutingCondition is a predicate that decides whether a given HTTP request
// should be routed to the associated handler. It returns true if the request
// matches the condition (e.g., path, method, or header), and false otherwise.
type RoutingCondition func(request *http.Request) bool

// Route pairs a RoutingCondition with an HTTP handler function. If the condition returns
// true for an incoming request, the handler is invoked by the webhook server.
type Route datautils.Pair[RoutingCondition, http.HandlerFunc]

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
	conn ConnectorWebhookSubscriber,
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
