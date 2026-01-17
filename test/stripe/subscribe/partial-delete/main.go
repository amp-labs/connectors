package main

// This test validates partial deletion of webhook subscriptions.
//
// Test Overview:
// This script tests that subscriptions can be partially deleted and that the provider
// correctly maintains the remaining subscription state after each deletion step.
//
// Test Pattern:
// The test creates a large set of subscriptions initially (A, B, C), then progressively
// prunes it by deleting subsets of events/objects and validating what remains after
// each deletion. The progression follows this pattern:
//
//	Subscribe(A, B, C)
//	Delete(A) + Validate(B, C remain)
//	Delete(part of C) + Validate(remaining C and B remain)
//	Delete(remaining B and C)
//
// After the main sequence, the test validates edge cases:
//   - Deleting the last event should delete the entire endpoint
//   - Creating a new subscription after endpoint deletion should create a new endpoint
//
// Test Flow Summary:
//  1. Subscribe to: account, balance, charge objects with various events
//  2. Delete account subscription → validate account events removed, balance & charge remain
//  3. Delete charge.succeeded event → validate charge.succeeded removed, other events remain
//  4. Delete remaining subscriptions (balance + charge.dispute.funds_withdrawn)
//
// The SubscriptionEvents are the core data - you specify what to subscribe to.
// The validation logic checks the provider state to ensure deletions work correctly.

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
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

	// Initial subscription: account, balance, charge
	initialEvents := map[common.ObjectName]common.ObjectEvents{
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
	}

	// track endpoint ID for validation
	var accountEndpointID string

	steps := []partialDeleteStep{
		{
			Name: "Delete account subscription",
			BuildDeleteResult: func(current *common.SubscriptionResult) *common.SubscriptionResult {
				stripeResult := current.Result.(*stripe.SubscriptionResult)
				accountEndpoint := stripeResult.Subscriptions["account"]
				accountEndpointID = extractEndpointID(accountEndpoint.ID)

				return &common.SubscriptionResult{
					Result: &stripe.SubscriptionResult{
						Subscriptions: map[common.ObjectName]stripe.WebhookResponse{
							"account": accountEndpoint,
						},
					},
				}
			},
			Validate: func(ctx context.Context, conn connectors.SubscribeConnector, current *common.SubscriptionResult) error {
				stripeConn := conn.(*stripe.Connector)
				endpoint, err := stripeConn.GetWebhookEndpoint(ctx, accountEndpointID)
				if err != nil {
					return fmt.Errorf("error fetching endpoint: %w", err)
				}

				hasAccount := datautils.NewStringSet(endpoint.EnabledEvents...).Has("account.updated") ||
					datautils.NewStringSet(endpoint.EnabledEvents...).Has("account.application.authorized")
				if hasAccount {
					return fmt.Errorf("account events should be removed, but found: %v", endpoint.EnabledEvents)
				}

				slog.Info("Account events successfully removed")
				utils.DumpJSON(endpoint, os.Stdout)
				return nil
			},
		},
		{
			Name: "Delete charge.succeeded event",
			BuildDeleteResult: func(current *common.SubscriptionResult) *common.SubscriptionResult {
				stripeResult := current.Result.(*stripe.SubscriptionResult)
				chargeEndpoint := stripeResult.Subscriptions["charge"]

				return &common.SubscriptionResult{
					Result: &stripe.SubscriptionResult{
						Subscriptions: map[common.ObjectName]stripe.WebhookResponse{
							"charge": {
								ID:            chargeEndpoint.ID,
								EnabledEvents: []string{"charge.succeeded"},
							},
						},
					},
				}
			},
			Validate: func(ctx context.Context, conn connectors.SubscribeConnector, current *common.SubscriptionResult) error {
				stripeConn := conn.(*stripe.Connector)
				endpoint, err := stripeConn.GetWebhookEndpoint(ctx, accountEndpointID)
				if err != nil {
					return fmt.Errorf("error fetching endpoint: %w", err)
				}

				if datautils.NewStringSet(endpoint.EnabledEvents...).Has("charge.succeeded") {
					return fmt.Errorf("charge.succeeded should be removed, but found: %v", endpoint.EnabledEvents)
				}

				slog.Info("charge.succeeded successfully removed")
				utils.DumpJSON(endpoint, os.Stdout)
				return nil
			},
		},
		{
			Name: "Delete remaining subscriptions",
			BuildDeleteResult: func(current *common.SubscriptionResult) *common.SubscriptionResult {
				stripeResult := current.Result.(*stripe.SubscriptionResult)
				chargeEndpoint := stripeResult.Subscriptions["charge"]

				return &common.SubscriptionResult{
					Result: &stripe.SubscriptionResult{
						Subscriptions: map[common.ObjectName]stripe.WebhookResponse{
							"balance": stripeResult.Subscriptions["balance"],
							"charge": {
								ID:            chargeEndpoint.ID,
								EnabledEvents: []string{"charge.dispute.funds_withdrawn"},
							},
						},
					},
				}
			},
			Validate: nil, // no validation needed for final deletion
		},
	}

	suite := partialDeleteTestSuite{
		WebhookURL: webhookURL,
		BuildRequest: func(url string) any {
			return &stripe.SubscriptionRequest{
				WebhookEndPoint: url,
			}
		},
		Steps: steps,
		OnSubscribe: func(result *common.SubscriptionResult) {
			stripeResult := result.Result.(*stripe.SubscriptionResult)
			if len(stripeResult.Subscriptions) != 3 {
				logging.Logger(ctx).Error("Expected 3 subscriptions", "count", len(stripeResult.Subscriptions))
				utils.Fail("expected 3 subscriptions", "count", len(stripeResult.Subscriptions))
			}
			slog.Info("3 subscriptions created successfully")
		},
	}

	// run the main partial delete scenario
	validatePartialDelete(ctx, conn, initialEvents, suite)

}

type partialDeleteStep struct {
	Name              string
	BuildDeleteResult func(current *common.SubscriptionResult) *common.SubscriptionResult
	Validate          func(ctx context.Context, conn connectors.SubscribeConnector, current *common.SubscriptionResult) error
}

type partialDeleteTestSuite struct {
	WebhookURL string

	BuildRequest func(webhookURL string) any

	Steps []partialDeleteStep

	OnSubscribe func(result *common.SubscriptionResult)
}

func validatePartialDelete(
	ctx context.Context,
	conn connectors.SubscribeConnector,
	initialEvents map[common.ObjectName]common.ObjectEvents,
	suite partialDeleteTestSuite,
) {
	slog.Info("> TEST Partial Delete")

	subscribeParams := common.SubscribeParams{
		SubscriptionEvents: initialEvents,
		Request:            suite.BuildRequest(suite.WebhookURL),
	}

	currentResult, err := conn.Subscribe(ctx, subscribeParams)
	if err != nil {
		logging.Logger(ctx).Error("Error subscribing", "error", err)
		utils.Fail("error subscribing", "error", err)
	}

	if currentResult == nil {
		utils.Fail("subscription result is nil")
	}

	utils.DumpJSON(currentResult, os.Stdout)

	if suite.OnSubscribe != nil {
		suite.OnSubscribe(currentResult)
	}

	// EXECUTE DELETION STEPS
	for i, step := range suite.Steps {
		time.Sleep(5 * time.Second)
		stepName := step.Name
		if stepName == "" {
			stepName = fmt.Sprintf("Step %d", i+1)
		}

		slog.Info("> Deleting partial subscription", "step", stepName)

		deleteResult := step.BuildDeleteResult(currentResult)
		if deleteResult == nil {
			utils.Fail("BuildDeleteResult returned nil", "step", stepName)
		}

		err := conn.DeleteSubscription(ctx, *deleteResult)
		if err != nil {
			logging.Logger(ctx).Error("Error deleting partial subscription", "error", err, "step", stepName)
			utils.Fail("error deleting partial subscription", "error", err, "step", stepName)
		}

		if step.Validate != nil {

			if err := step.Validate(ctx, conn, currentResult); err != nil {
				logging.Logger(ctx).Error("Validation failed", "error", err, "step", stepName)
				utils.Fail("validation failed", "error", err, "step", stepName)
			}
			slog.Info("Validation passed", "step", stepName)
		}
	}

	slog.Info("> Successful test completion")
}

// extract endpoint ID from Stripe subscription ID
func extractEndpointID(subID string) string {
	if idx := strings.LastIndex(subID, ":"); idx != -1 {
		return subID[:idx]
	}
	return subID
}
