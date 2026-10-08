package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/amp-common/webhook"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/subscribe/deps"
)

var (
	errAttioWebhookSecretNotFound = errors.New("attio webhook secret not found for installation")

	// errSubscriptionListerNotConfigured is returned when the Attio verification-params builder
	// is invoked without a deps.Dependencies.Subscriptions resolver — the stored webhook signing
	// secret cannot be recovered without it.
	errSubscriptionListerNotConfigured = errors.New("subscription result lister not configured in dependencies")
)

// attioConfig is the per-provider subscribe-config bundle for Attio. Attio subscribes via API and
// builds a subscribe-time request payload (the webhook endpoint), so a buildRequestFn is declared.
// Its webhook payload is object-shaped ({webhook_id, events: [...]}) and fanned out into individual
// events by attio.CollapsedSubscriptionEvent (registered in GetObjectTypeSubscribeEventsList).
//
// Webhook signatures are verified: Attio signs each webhook with a secret it generates at
// webhook-creation time and returns in the SubscriptionResult. getAttioVerificationParams recovers
// that stored secret (via deps.Dependencies.Subscriptions) so the connector's VerifyWebhookMessage can
// HMAC-check the request.
var attioConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getAttioRequest,
		// Attio webhooks do not report which fields changed (SubscriptionEvent.UpdatedFields
		// is empty), so field-level watching is impossible; Attio requires WatchFieldsAuto="all".
		requiresWatchFieldsAuto: true,
	},
	Verification: VerificationConfig{
		paramsFn:          getAttioVerificationParams,
		verifierConnector: &attio.Connector{},
	},
}

func getAttioRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &attio.SubscriptionRequest{
		WebhookEndpoint: webhookURL,
	}, nil
}

// getAttioVerificationParams supplies the secret Attio generated for the installation's webhook so
// the connector can verify the incoming signature.
func getAttioVerificationParams(
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

	secret, err := attioWebhookSecret(ctx, deps, req.Installation.Id, req.Payload)
	if err != nil {
		return nil, err
	}

	return &common.VerificationParams{
		Param: &attio.AttioVerificationParams{Secret: secret},
	}, nil
}

// attioWebhookSecret returns the signing secret for the webhook that produced the incoming payload.
// Attio signs each webhook with its own generated secret (stored in the SubscriptionResult at
// creation time), so we match the payload's webhook_id to the stored subscription and return its
// secret. When only one subscription exists we fall back to it even if the id is unavailable.
func attioWebhookSecret(
	ctx context.Context, deps deps.Dependencies, installationID string, payload *webhook.PubsubPayload,
) (string, error) {
	webhookID := parseAttioWebhookID(payload)

	results, err := deps.Subscriptions.ListSubscriptionResults(ctx, installationID,
		func() *common.SubscriptionResult {
			return &common.SubscriptionResult{Result: &attio.SubscriptionResult{}}
		})
	if err != nil {
		return "", fmt.Errorf("error listing subscription results for installation %q: %w", installationID, err)
	}

	var fallback string

	for _, full := range results {
		if full == nil {
			continue
		}

		result, ok := full.Result.(*attio.SubscriptionResult)
		if !ok || result.Data.Secret == "" {
			continue
		}

		if webhookID != "" && result.Data.Id.WebhookId == webhookID {
			return result.Data.Secret, nil
		}

		if fallback == "" {
			fallback = result.Data.Secret
		}
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("%w: %q", errAttioWebhookSecretNotFound, installationID)
}

// parseAttioWebhookID extracts the top-level webhook_id from a raw Attio webhook body. Returns an
// empty string when the body is missing or unparsable.
func parseAttioWebhookID(payload *webhook.PubsubPayload) string {
	if payload == nil {
		return ""
	}

	var body struct {
		WebhookID string `json:"webhook_id"`
	}

	err := json.Unmarshal(payload.Message, &body)
	if err != nil {
		return ""
	}

	return body.WebhookID
}
