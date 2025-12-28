package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/internal/datautils"
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

	subscriptionResult := subscribeResult.Result.(*stripe.SubscriptionResult)
	if len(subscriptionResult.Subscriptions) != 3 {
		logging.Logger(ctx).Error("Expected 3 subscriptions", "count", len(subscriptionResult.Subscriptions))
		return
	}
	slog.Info("Created 3 subscriptions successfully")
	utils.DumpJSON(subscribeResult, os.Stdout)

	slog.Info("Deleting account subscription...")

	partialDeleteResult := common.SubscriptionResult{
		Result: &stripe.SubscriptionResult{
			Subscriptions: map[common.ObjectName]stripe.WebhookEndpointResponse{
				"account": subscribeResult.Result.(*stripe.SubscriptionResult).Subscriptions["account"],
			},
		},
	}

	err = conn.DeleteSubscription(ctx, partialDeleteResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting account subscription", "error", err)
		return
	}

	slog.Info("Verifying remaining subscriptions...")

	accountEndpoint := subscriptionResult.Subscriptions["account"]
	realEndpointID := accountEndpoint.ID
	if idx := strings.LastIndex(realEndpointID, ":"); idx != -1 {
		realEndpointID = realEndpointID[:idx]
	}

	endpoint, err := conn.GetWebhookEndpoint(ctx, realEndpointID)
	if err != nil {
		logging.Logger(ctx).Error("Error fetching endpoint", "error", err)
		return
	}

	hasAccount := datautils.NewStringSet(endpoint.EnabledEvents...).Has("account.updated") || datautils.NewStringSet(endpoint.EnabledEvents...).Has("account.application.authorized")
	if hasAccount {
		logging.Logger(ctx).Error("✗ Account events should be removed", "events", endpoint.EnabledEvents)
		return
	}
	slog.Info("Account events successfully removed")
	utils.DumpJSON(endpoint, os.Stdout)

	slog.Info("Deleting charge.succeeded event...")

	chargeEndpoint := subscriptionResult.Subscriptions["charge"]

	partialChargeDeleteResult := common.SubscriptionResult{
		Result: &stripe.SubscriptionResult{
			Subscriptions: map[common.ObjectName]stripe.WebhookEndpointResponse{
				"charge": {
					ID:            chargeEndpoint.ID,
					EnabledEvents: []string{"charge.succeeded"},
				},
			},
		},
	}

	err = conn.DeleteSubscription(ctx, partialChargeDeleteResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting charge.succeeded event", "error", err)
		return
	}

	slog.Info("Verifying remaining events...")

	endpoint, err = conn.GetWebhookEndpoint(ctx, realEndpointID)
	if err != nil {
		logging.Logger(ctx).Error("Error fetching endpoint", "error", err)
		return
	}

	if datautils.NewStringSet(endpoint.EnabledEvents...).Has("charge.succeeded") {
		logging.Logger(ctx).Error("✗ charge.succeeded should be removed", "events", endpoint.EnabledEvents)
		return
	}
	slog.Info("charge.succeeded successfully removed")
	utils.DumpJSON(endpoint, os.Stdout)

	slog.Info("Deleting remaining subscriptions...")

	remainingDeleteResult := common.SubscriptionResult{
		Result: &stripe.SubscriptionResult{
			Subscriptions: map[common.ObjectName]stripe.WebhookEndpointResponse{
				"balance": subscriptionResult.Subscriptions["balance"],
				"charge": {
					ID:            chargeEndpoint.ID,
					EnabledEvents: []string{"charge.dispute.funds_withdrawn"},
				},
			},
		},
	}

	err = conn.DeleteSubscription(ctx, remainingDeleteResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting remaining subscriptions", "error", err)
		return
	}

	slog.Info("Testing deletion of last event (should delete entire endpoint)...")

	newSubscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"balance": {
				PassThroughEvents: []string{"balance.available"},
			},
		},
		Request: &stripe.SubscriptionRequest{
			WebhookEndPoint: webhookURL,
		},
	}

	newSubscribeResult, err := conn.Subscribe(ctx, newSubscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error creating new subscription", "error", err)
		return
	}

	newSubscriptionResult := newSubscribeResult.Result.(*stripe.SubscriptionResult)
	if len(newSubscriptionResult.Subscriptions) != 1 {
		logging.Logger(ctx).Error("✗ Expected 1 subscription", "count", len(newSubscriptionResult.Subscriptions))
		return
	}
	slog.Info("Created new subscription successfully")
	utils.DumpJSON(newSubscribeResult, os.Stdout)

	newBalanceEndpoint := newSubscriptionResult.Subscriptions["balance"]
	newEndpointID := newBalanceEndpoint.ID
	if idx := strings.LastIndex(newEndpointID, ":"); idx != -1 {
		newEndpointID = newEndpointID[:idx]
	}

	slog.Info("Deleting last remaining event (balance.available)...")

	lastEventDeleteResult := common.SubscriptionResult{
		Result: &stripe.SubscriptionResult{
			Subscriptions: map[common.ObjectName]stripe.WebhookEndpointResponse{
				"balance": newBalanceEndpoint,
			},
		},
	}

	err = conn.DeleteSubscription(ctx, lastEventDeleteResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting last event", "error", err)
		return
	}

	slog.Info("Verifying endpoint was deleted...")

	_, err = conn.GetWebhookEndpoint(ctx, newEndpointID)
	if err == nil {
		logging.Logger(ctx).Error("✗ Endpoint still exists after deleting last event")
		return
	}
	slog.Info("Endpoint successfully deleted (as expected)")

	slog.Info("Creating new subscription after deletion (should create new endpoint)...")

	finalSubscribeParams := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"charge": {
				PassThroughEvents: []string{"charge.succeeded"},
			},
		},
		Request: &stripe.SubscriptionRequest{
			WebhookEndPoint: webhookURL,
		},
	}

	finalSubscribeResult, err := conn.Subscribe(ctx, finalSubscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error creating final subscription", "error", err)
		return
	}

	finalSubscriptionResult := finalSubscribeResult.Result.(*stripe.SubscriptionResult)
	if len(finalSubscriptionResult.Subscriptions) != 1 {
		logging.Logger(ctx).Error("✗ Expected 1 subscription", "count", len(finalSubscriptionResult.Subscriptions))
		return
	}
	slog.Info("Created final subscription successfully")
	utils.DumpJSON(finalSubscribeResult, os.Stdout)

	finalChargeEndpoint := finalSubscriptionResult.Subscriptions["charge"]
	finalEndpointID := finalChargeEndpoint.ID
	if idx := strings.LastIndex(finalEndpointID, ":"); idx != -1 {
		finalEndpointID = finalEndpointID[:idx]
	}

	slog.Info("Verifying new endpoint was created...")

	finalEndpoint, err := conn.GetWebhookEndpoint(ctx, finalEndpointID)
	if err != nil {
		logging.Logger(ctx).Error("✗ Error fetching new endpoint", "error", err)
		return
	}
	slog.Info("New endpoint verified")
	utils.DumpJSON(finalEndpoint, os.Stdout)

	if finalEndpointID == newEndpointID {
		logging.Logger(ctx).Warn("New endpoint has same ID as deleted endpoint", "endpointID", finalEndpointID)
	} else {
		slog.Info("New endpoint created with different ID", "oldEndpointID", newEndpointID, "newEndpointID", finalEndpointID)
	}

	slog.Info("Cleaning up - deleting final subscription...")

	err = conn.DeleteSubscription(ctx, *finalSubscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting final subscription", "error", err)
		return
	}

	slog.Info("Final subscription deleted successfully")

	slog.Info("Partial delete test completed")
}
