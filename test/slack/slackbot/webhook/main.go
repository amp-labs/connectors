package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := slackshared.NewConnector(ctx, providers.Slack)

	testscenario.RunWebhookConsumer(ctx, slackshared.NewWebhookProcessor(), conn, nil)
}
