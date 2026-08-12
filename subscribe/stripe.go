package subscribe

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// errStripeWebhookSecretNotFound is returned when no stored subscription result carries the
// endpoint signing secret needed to verify an incoming Stripe webhook.
var errStripeWebhookSecretNotFound = errors.New("stripe webhook signing secret not found for installation")

// stripeConfig is the per-provider subscribe-config bundle for Stripe. Stripe subscribes via
// API and builds a subscribe-time request payload (the webhook endpoint URL), so a
// buildRequestFn is declared. Stripe sends one event per webhook delivery; the payload is
// fanned out by stripe.CollapsedSubscriptionEvent (registered in
// GetObjectTypeSubscribeEventsList).
//
// Webhook signatures are verified: Stripe generates a signing secret when the webhook
// endpoint is created and the connector stores it in the SubscriptionResult.
// getStripeVerificationParams recovers that stored secret (via
// deps.Dependencies.Subscriptions) so the connector's VerifyWebhookMessage can HMAC-check
// the Stripe-Signature header.
var stripeConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getStripeRequest,
	},
	Verification: VerificationConfig{
		paramsFn:          getStripeVerificationParams,
		verifierConnector: &stripe.Connector{},
		eventCaster:       CastSubscriptionEvents[stripe.SubscriptionEvent],
	},
}

func getStripeRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &stripe.SubscriptionRequest{
		WebhookEndPoint: webhookURL,
	}, nil
}

// getStripeVerificationParams supplies the signing secret Stripe generated for the
// installation's webhook endpoint so the connector can verify the incoming signature.
func getStripeVerificationParams(
	ctx context.Context,
	deps deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	if deps.Subscriptions == nil {
		return nil, errSubscriptionListerNotConfigured
	}

	secret, err := stripeWebhookSecret(ctx, deps, req.Installation.Id)
	if err != nil {
		return nil, err
	}

	return &common.VerificationParams{
		Param: &stripe.VerificationParams{Secret: secret},
	}, nil
}

// stripeWebhookSecret returns the signing secret for the installation's webhook endpoint.
// Stripe returns the secret once, at endpoint-creation time, and the connector persists it
// in the SubscriptionResult; all of an installation's object subscriptions share a single
// endpoint (composite IDs "endpointID:objectName"), so the first stored secret applies.
func stripeWebhookSecret(
	ctx context.Context, deps deps.Dependencies, installationID string,
) (string, error) {
	results, err := deps.Subscriptions.ListSubscriptionResults(ctx, installationID,
		func() *common.SubscriptionResult {
			return &common.SubscriptionResult{Result: &stripe.SubscriptionResult{}}
		})
	if err != nil {
		return "", fmt.Errorf("error listing subscription results for installation %q: %w", installationID, err)
	}

	for _, full := range results {
		if full == nil {
			continue
		}

		result, ok := full.Result.(*stripe.SubscriptionResult)
		if !ok {
			continue
		}

		for _, endpoint := range result.Subscriptions {
			if endpoint.Secret != "" {
				return endpoint.Secret, nil
			}
		}
	}

	return "", fmt.Errorf("%w: %q", errStripeWebhookSecretNotFound, installationID)
}
