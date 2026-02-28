package testscenario

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/test/utils"
)

// SubscribeTestSuite controls the ValidateSubscribeUpdateDelete procedure.
type SubscribeTestSuite struct {
	// WebhookURL is the URL where webhooks will be sent. Required.
	WebhookURL string

	// BuildRequest is a function that creates a provider-specific Request object
	// for SubscribeParams. The function receives the webhookURL as a parameter.
	// For example, for Stripe: func(url string) any { return &stripe.SubscriptionRequest{WebhookEndPoint: url} }
	BuildRequest func(webhookURL string) any

	// OnSubscribe is an optional callback invoked after successful subscription
	// with the subscription result for custom processing.
	OnSubscribe func(result *common.SubscriptionResult)

	// OnUpdate is an optional callback invoked after successful update
	// with the update result for custom processing.
	OnUpdate func(result *common.SubscriptionResult)

	// OnDelete is an optional callback invoked before deletion.
	OnDelete func(result *common.SubscriptionResult)
}

// ValidateSubscribeUpdateDelete performs a comprehensive subscription test:
// Subscribe -> Update -> Delete.
//
// Flow:
// 1. Subscribe with initial SubscriptionEvents
// 2. Update subscription with new SubscriptionEvents
// 3. Delete subscription at the end
//
// The SubscriptionEvents are the core data. The rest (Subscribe/Update/Delete operations)
// is handled mechanically using the SubscribeConnector interface.
func ValidateSubscribeUpdateDelete(
	ctx context.Context,
	conn connectors.SubscribeConnector,
	initialEvents map[common.ObjectName]common.ObjectEvents,
	updateEvents map[common.ObjectName]common.ObjectEvents,
	suite SubscribeTestSuite,
) {
	slog.Info("> TEST Subscribe/Update/Delete")

	// SUBSCRIBE
	slog.Info("Creating subscriptions...")
	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: initialEvents,
		Request:            suite.BuildRequest(suite.WebhookURL),
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		utils.Fail("error subscribing", "error", err)
	}

	if subscribeResult == nil {
		utils.Fail("subscription result is nil")
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	if suite.OnSubscribe != nil {
		suite.OnSubscribe(subscribeResult)
	}

	// UPDATE
	slog.Info("Updating subscriptions...")
	updateParams := common.SubscribeParams{
		SubscriptionEvents: updateEvents,
		Request:            suite.BuildRequest(suite.WebhookURL),
	}

	updateResult, err := conn.UpdateSubscription(ctx, updateParams, subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error updating subscription", "error", err)
		utils.Fail("error updating subscription", "error", err)
	}

	if updateResult == nil {
		utils.Fail("update result is nil")
	}

	utils.DumpJSON(updateResult, os.Stdout)

	if suite.OnUpdate != nil {
		suite.OnUpdate(updateResult)
	}

	// DELETE
	slog.Info("Deleting subscriptions...")
	if suite.OnDelete != nil {
		suite.OnDelete(updateResult)
	}

	err = conn.DeleteSubscription(ctx, *updateResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)
		utils.Fail("error deleting subscription", "error", err)
	}

	slog.Info("Delete subscription successful")
	slog.Info("> Successful test completion")
}

// ValidateSubscribeDelete performs a subscription test: Subscribe -> Delete.
//
// Flow:
// 1. Subscribe with SubscriptionEvents
// 2. Delete subscription at the end
//
// The SubscriptionEvents are the core data. The rest (Subscribe/Delete operations)
// is handled mechanically using the SubscribeConnector interface.
func ValidateSubscribeDelete(
	ctx context.Context,
	conn connectors.SubscribeConnector,
	events map[common.ObjectName]common.ObjectEvents,
	suite SubscribeTestSuite,
) {
	slog.Info("> TEST Subscribe/Delete")

	// SUBSCRIBE
	slog.Info("Creating subscriptions...")
	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: events,
		Request:            suite.BuildRequest(suite.WebhookURL),
	}

	subscribeResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		utils.Fail("error subscribing", "error", err)
	}

	if subscribeResult == nil {
		utils.Fail("subscription result is nil")
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	if suite.OnSubscribe != nil {
		suite.OnSubscribe(subscribeResult)
	}

	// DELETE
	slog.Info("Deleting subscriptions...")
	if suite.OnDelete != nil {
		suite.OnDelete(subscribeResult)
	}

	err = conn.DeleteSubscription(ctx, *subscribeResult)
	if err != nil {
		logging.Logger(ctx).Error("Error deleting subscription", "error", err)
		utils.Fail("error deleting subscription", "error", err)
	}

	slog.Info("Delete subscription successful")
	slog.Info("> Successful test completion")
}
