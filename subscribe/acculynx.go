package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// acculynxTechContact is the email AccuLynx contacts about subscription status. AccuLynx requires a
// valid email on every subscription, and there is no per-installation/consumer email available in
// the subscribe flow, so Ampersand registers its own address centrally.
const acculynxTechContact = "support@withampersand.com"

// acculynxConfig is the per-provider subscribe-config bundle for AccuLynx. AccuLynx subscribes via
// API and builds a subscribe-time request payload (webhook consumer URL + tech-contact email), so a
// buildRequestFn is declared. Webhook signature verification is not yet implemented on AccuLynx (the
// connector's VerifyWebhookMessage is a placeholder that accepts all), so verification is bypassed;
// an event caster unwraps its webhook payloads into typed SubscriptionEvents.
var acculynxConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getAccuLynxRequest,
	},
	Verification: VerificationConfig{
		verifierConnector: &acculynx.Connector{},
		bypassed:          true,
		eventCaster:       castSubscriptionEvents[acculynx.SubscriptionEvent],
	},
}

func getAccuLynxRequest(
	_ context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	return &acculynx.SubscriptionRequest{
		ConsumerURL: webhookURL,
		TechContact: acculynxTechContact,
	}, nil
}
