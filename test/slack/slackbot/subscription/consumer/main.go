package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/slack/slackbot"
	"github.com/amp-labs/connectors/test/slack/slackbot/subscription"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := slackbot.NewConnector(ctx)

	testscenario.RunWebhookConsumer(ctx, subscription.NewWebhookProcessor(), conn, nil)
}
