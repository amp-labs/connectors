package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	webhookURL := "https://play.svix.com/in/e_Tqq6rWtd3gm9urmVm7zu7OBT6aW"

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
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
		},
		Request: &stripe.SubscriptionRequest{
			WebhookEndPoint: webhookURL,
		},
	}

	slog.Info("Creating subscriptions...")
	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	slog.Info("Updating subscriptions...")
	updateParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"account": {
				Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
				PassThroughEvents: []string{"account.application.authorized"},
			},
			"charge": {
				PassThroughEvents: []string{"charge.succeeded"},
			},
		},
		Request: &stripe.SubscriptionRequest{
			WebhookEndPoint: webhookURL,
		},
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err)
		return
	}

	utils.DumpJSON(updateResult, os.Stdout)

	slog.Info("Deleting subscriptions...")
	err = conn.DeleteSubscription(ctx, *subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)
		return
	}

	slog.Info("Delete subscription successful")
}
