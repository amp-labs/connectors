package subscribe

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// RegistrationConfig holds the per-provider declarative configuration for the one-time
// registration step that some providers require before per-object subscriptions can be created
// (e.g. Salesforce → AWS EventBridge wiring).
//
// Per-provider files declare a RegistrationConfig literal populating only the fields that
// are *not* derivable from ProviderInfo (currently emptyResultFn and buildParamsFn). The
// "does this provider require registration?" answer is computed from ProviderInfo by
// IsRequired and is therefore not stored on the struct. The module / providerInfo fields are
// bound at GetProviderConfig time so the methods can answer module-aware questions.
//
// The zero RegistrationConfig is valid: methods on it are safe no-ops, returning false /
// nil / ErrRegistrationParamsBuilderNotDeclared as appropriate. ProviderConfig embeds this
// type by value, and the methods themselves take a value receiver, so callers can invoke
// them directly without nil-checking — the type system guarantees a non-nil receiver.
type RegistrationConfig struct {
	// emptyResultFn returns a fresh, mutable *common.RegistrationResult with the provider-
	// specific .Result struct populated as a zero value. Required (rather than a singleton)
	// because downstream code mutates .Result as the registration progresses — sharing a single
	// instance would race across concurrent registrations.
	emptyResultFn func() *common.RegistrationResult

	// buildParamsFn constructs the provider-specific request payload that goes inside
	// common.SubscriptionRegistrationParams.Request for a given installation.
	buildParamsFn func(ctx context.Context, inst *openapi.Installation) (any, error)

	// module and providerInfo are bound by GetProviderConfig at call time. IsRequired reads
	// them to compute the answer from ProviderInfo.
	module       common.ModuleID
	providerInfo *providers.ProviderInfo
}

// IsRequired reports whether a registration should be created for this provider given the
// supplied connector:
//
//  1. The provider's ProviderInfo indicates that registration is required.
//  2. The connector is non-nil.
//  3. The connector implements connectors.RegisterSubscribeConnector.
//
// The connector parameter is passed in at call time because connectors are not always
// available when ProviderConfig is fetched.
func (r RegistrationConfig) IsRequired(connector connectors.SubscribeConnector) bool {
	return shouldCreateRegistration(r.module, r.providerInfo, connector)
}

// EmptyResult returns a fresh empty registration result instance with the appropriate provider-
// specific result type, or nil if the provider does not declare an empty-result factory.
func (r RegistrationConfig) EmptyResult() *common.RegistrationResult {
	if r.emptyResultFn == nil {
		return nil
	}

	return r.emptyResultFn()
}

// BuildParams invokes the provider-specific registration params builder.
//
// Behavior when buildParamsFn is not declared:
//   - If the provider does not require registration (per ProviderInfo), returns
//     ErrRegistrationParamsBuilderNotDeclared. Calling BuildParams in this case is a bug —
//     the provider isn't a registration target at all.
//   - If the provider does require registration but declared no builder, returns (nil, nil).
//     This supports providers whose connector.Register accepts an empty / nil Request.
func (r RegistrationConfig) BuildParams(ctx context.Context, inst *openapi.Installation) (any, error) {
	if r.buildParamsFn != nil {
		return r.buildParamsFn(ctx, inst)
	}

	if !providerRequiresRegistration(r.module, r.providerInfo) {
		return nil, ErrRegistrationParamsBuilderNotDeclared
	}

	return nil, nil //nolint:nilnil // documented contract: registration required but no custom Request payload.
}
