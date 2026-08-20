package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/slack/slackuser"
	"github.com/amp-labs/connectors/test/slack/slackuser/subscription"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := slackuser.NewConnector(ctx)

	testscenario.RunWebhookConsumer(ctx, subscription.NewWebhookProcessor(), conn, nil)
}
