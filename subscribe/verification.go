package subscribe

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// verificationParamsFunc is a function that retrieves webhook verification parameters for a
// specific provider. It takes the bound deps.Dependencies plus the deps.VerificationRequest (payload,
// integration, installation, provider app details) and returns the verification parameters
// needed to validate incoming webhook requests from that provider.
type verificationParamsFunc func(
	ctx context.Context,
	deps deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error)

var (
	errVerificationParamsFuncNotFound  = errors.New("no verification params func found for provider")
	errWebhookVerificationNotSupported = errors.New("webhook verification not supported")
)

// isHookdeckGatewayProvider reports whether the given provider's webhooks are routed through
// Hookdeck. Used both to gate Hookdeck signature verification and to choose between Hookdeck and
// CloudFunction event endpoints.
//
// Salesforce / SalesforceJWT use a CloudFunction endpoint with their own AWS EventBridge wiring,
// and Hubspot delivers webhooks directly. Everyone else routes through Hookdeck.
//
// TODO should this be preserved?
func isHookdeckGatewayProvider(provider providers.Provider) bool { // nolint:unused
	return provider != providers.Hubspot &&
		provider != providers.Salesforce &&
		provider != providers.SalesforceJWT
}

// VerificationConfig is the subcomponent of ProviderConfig that exposes the webhook-verification
// and event-receipt concerns: building per-provider verification params, obtaining the webhook-
// verifier connector, whether signature verification is bypassed, and casting raw event maps into
// typed SubscriptionEvents.
//
// All four are non-derivable data (no ProviderInfo equivalent), so they are declared
// per-provider — there is no derived method here, hence no bound module/providerInfo. The zero
// VerificationConfig is valid; ProviderConfig embeds it by value, and the methods take a value
// receiver, so callers can invoke them directly without nil-checking. A nil paramsFn /
// verifierConnector / eventCaster means the provider has no such concern (the corresponding
// method returns the matching sentinel error); bypassed defaults to false.
type VerificationConfig struct {
	// paramsFn builds the provider-specific webhook verification params. Nil when the provider
	// has none (e.g. Gong, HousecallPro).
	paramsFn verificationParamsFunc

	// verifierConnector is the connector instance used to verify incoming webhook signatures.
	// Nil when the provider has no verifier connector. Shared across calls — a zero-value
	// connector exposing only read-only verification methods. Callers wanting instrumentation
	// (e.g. metrics) wrap the returned connector themselves.
	verifierConnector connectors.WebhookVerifierConnector

	// bypassed reports whether webhook signature verification is skipped for this provider.
	bypassed bool

	// eventCaster casts raw event maps into typed SubscriptionEvents. Nil when the provider's
	// events need no provider-specific casting.
	eventCaster subscriptionEventCaster

	// deps is bound by GetProviderConfig at call time and threaded into paramsFn.
	deps deps.Dependencies
}

// Params builds the provider's webhook verification params. Returns
// errVerificationParamsFuncNotFound when the provider declares no paramsFn.
func (v VerificationConfig) Params(
	ctx context.Context,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if v.paramsFn == nil {
		return nil, errVerificationParamsFuncNotFound
	}

	return v.paramsFn(ctx, v.deps, req)
}

// Connector returns the webhook-verifier connector. Returns
// errWebhookVerificationNotSupported when the provider declares no verifierConnector.
//
// The connector is returned unwrapped; callers wanting instrumentation (e.g. metrics) decorate
// it themselves. The context is accepted (though unused here) so the signature matches the
// original server implementation this package was migrated from — call sites and downstream
// decorators keep their ctx-passing shape.
func (v VerificationConfig) Connector(_ context.Context) (connectors.WebhookVerifierConnector, error) {
	if v.verifierConnector == nil {
		return nil, errWebhookVerificationNotSupported
	}

	return v.verifierConnector, nil
}

// ShouldBypass reports whether webhook signature verification should be skipped for this provider.
func (v VerificationConfig) ShouldBypass() bool {
	return v.bypassed
}

// CastEvents casts raw event maps into the provider's typed SubscriptionEvents. Returns
// errSubscriptionEventCasterNotDeclared when the provider declares no eventCaster.
func (v VerificationConfig) CastEvents(list []map[string]any) ([]common.SubscriptionEvent, error) {
	if v.eventCaster == nil {
		return nil, errSubscriptionEventCasterNotDeclared
	}

	return v.eventCaster(list)
}
