package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// subscriptionRequestBuilder constructs the provider-specific subscribe-time request payload
// (e.g. Pub/Sub topic name for Google, webhook URL for Outreach/Salesloft/Zoho) that is passed
// to the connector's Subscribe method.
//
// inst / rev / conn are the shared amp-common wire types; conn is the same value as
// inst.Connection. deps carries the caller-provided resolver capabilities (bound at
// GetProviderConfig time). webhookURL is the event-receipt endpoint constructed by the caller;
// providers that do not need it (e.g. Google's Pub/Sub topic builder) simply ignore it.
type subscriptionRequestBuilder func(
	ctx context.Context,
	deps deps.Dependencies,
	inst *openapi.Installation,
	rev *openapi.Revision,
	registrationResult *common.RegistrationResult,
	conn *openapi.Connection,
	webhookURL string,
) (any, error)

// SubscriptionConfig is the subcomponent of ProviderConfig that exposes subscribe-step
// queries: programmatic-API support, look-up-only mode, the WatchFieldsAuto Salesloft quirk,
// and the subscribe-time request payload builder.
//
// The derivable answers (IsSupportedViaAPI, SubscribeManually) are computed from ProviderInfo —
// they are not stored. The non-derivable pieces are declared per-provider: buildRequestFn (the
// subscribe-time payload builder) and requiresWatchFieldsAuto (the Salesloft quirk; no
// ProviderInfo equivalent today).
//
// module / providerInfo / deps are bound at GetProviderConfig time. The zero SubscriptionConfig
// is valid; ProviderConfig embeds it by value, and the methods themselves take a value receiver,
// so callers can invoke them directly without nil-checking — the type system guarantees a
// non-nil receiver. A nil buildRequestFn means the provider needs no custom subscribe payload
// (BuildRequest returns nil, nil).
type SubscriptionConfig struct {
	// requiresWatchFieldsAuto indicates the provider requires WatchFieldsAuto="all" for
	// subscribe update events. Declared per-provider because there is no ProviderInfo
	// equivalent today.
	requiresWatchFieldsAuto bool

	// buildRequestFn constructs the subscribe-time request payload passed to the connector's
	// Subscribe method. Nil when the provider needs no custom payload.
	buildRequestFn subscriptionRequestBuilder

	// module, providerInfo, and deps are bound by GetProviderConfig at call time.
	module       common.ModuleID
	providerInfo *providers.ProviderInfo
	deps         deps.Dependencies
}

// IsSupportedViaAPI reports whether the provider supports programmatic subscription via API.
// Derived from ProviderInfo via subscriptionViaApiSupported.
func (s SubscriptionConfig) IsSupportedViaAPI() bool {
	return subscriptionViaApiSupported(s.module, s.providerInfo)
}

// SubscribeManually reports whether the provider's webhook subscriptions are configured manually
// at the provider level (rather than registered programmatically via API by Ampersand). Derived
// from ProviderInfo via isProviderLookUpOnly.
func (s SubscriptionConfig) SubscribeManually() bool {
	return isProviderLookUpOnly(s.module, s.providerInfo)
}

// RequiresWatchFieldsAutoAll reports whether the provider requires WatchFieldsAuto="all" for
// subscribe update events. Reads the per-provider declarative field.
func (s SubscriptionConfig) RequiresWatchFieldsAutoAll() bool {
	return s.requiresWatchFieldsAuto
}

// BuildRequest invokes the provider's subscribe-time request builder, passing the bound deps
// and &inst.Connection as the conn argument. webhookURL is the event-receipt endpoint
// constructed by the caller; builders that do not need it ignore it. Returns (nil, nil) when the
// provider declares no buildRequestFn — the provider needs no custom subscribe payload.
func (s SubscriptionConfig) BuildRequest(
	ctx context.Context,
	inst *openapi.Installation,
	rev *openapi.Revision,
	registrationResult *common.RegistrationResult,
	webhookURL string,
) (any, error) {
	if s.buildRequestFn == nil {
		return nil, nil //nolint:nilnil // documented contract: no builder → no custom subscribe payload.
	}

	// A declared builder always needs the installation: BuildRequest passes its Connection and
	// the per-provider builders read inst fields. Guard here so a nil installation is a clear
	// error rather than a panic.
	if inst == nil {
		return nil, errInstallationNotFound
	}

	return s.buildRequestFn(ctx, s.deps, inst, rev, registrationResult, &inst.Connection, webhookURL)
}
