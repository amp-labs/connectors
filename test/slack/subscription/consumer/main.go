package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/slack/subscription"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.NewConnector(ctx)

	testscenario.RunWebhookConsumer(ctx, conn, subscription.NewWebhookRouter(), nil)
}
