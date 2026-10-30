// Live webhook consumer: prints every Reply delivery that reaches the local server.
//
// Reply deliveries carry no signature, so verification accepts all of them and only the
// event shape is under test. Subscribe is not implemented at this stage, so create the
// subscription by hand once the public URL is known:
//
//  1. ngrok http 4550
//  2. go run ./test/reply/webhook, and give it the ngrok URL when asked
//  3. POST /v3/webhooks {"eventType": "contact_opted_out", "url": "<ngrok url>", "scope": "personal"}
//  4. POST /v3/webhooks/{id}/test sends a sample payload
//  5. DELETE /v3/webhooks/{id} when done
package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetReplyConnector(ctx)

	testscenario.RunWebhookConsumer(ctx, &testscenario.WebhookProcessor{}, conn, nil)
}
