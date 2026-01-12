package main

import (
	"context"
	"log/slog"
	"os"
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
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	webhookURLField := credscanning.Field{
		Name:      "webhookURL",
		PathJSON:  "webhookURL",
		SuffixENV: "WEBHOOK_URL",
	}
	filePath := credscanning.LoadPath(providers.Stripe)
	reader := utils.MustCreateProvCredJSON(filePath, false, webhookURLField)
	webhookURL := reader.Get(webhookURLField)

	if webhookURL == "" {
		slog.Error("Webhook URL is required. Add 'webhookURL' field to stripe credentials JSON file")
		os.Exit(1)
	}

	initialEvents := map[common.ObjectName]common.ObjectEvents{
		"account": {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			PassThroughEvents: []string{"account.application.authorized", "account.application.deauthorized"},
		},
		"balance": {
			PassThroughEvents: []string{"balance.available"},
		},
		"billing_portal": {
			PassThroughEvents: []string{"billing_portal.configuration.created"},
		},
		"charge": {
			PassThroughEvents: []string{"charge.dispute.funds_withdrawn", "charge.succeeded"},
		},
	}

	updateEvents := map[common.ObjectName]common.ObjectEvents{
		"account": {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			PassThroughEvents: []string{"account.application.authorized"},
		},
		"charge": {
			PassThroughEvents: []string{"charge.succeeded"},
		},
	}

	suite := testscenario.SubscribeTestSuite{
		WebhookURL: webhookURL,
		BuildRequest: func(url string) any {
			return &stripe.SubscriptionRequest{
				WebhookEndPoint: url,
			}
		},
	}

	testscenario.ValidateSubscribeUpdateDelete(ctx, conn, initialEvents, updateEvents, suite)
}
