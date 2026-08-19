package subscribe

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// ProviderConfig is the per-provider declarative bundle of subscribe-related configuration.
//
// Contributor model: to add subscribe support for a new provider, declare a literal of this
// type in subscribe/<provider>.go (for example var fooConfig = ProviderConfig{...}) and register
// it in providerConfigs below. The set of subcomponent fields a provider populates determines
// which subscribe-related operations are supported.
type ProviderConfig struct {
	// Registration holds the provider's one-time registration step configuration. Always
	// populated as a value (never nil) so callers can invoke its methods directly without
	// nil-checking. A provider that does not declare registration leaves the field at its
	// zero value; the zero RegistrationConfig's methods are safe no-ops.
	Registration RegistrationConfig

	// Subscription exposes subscribe-step queries (programmatic-API support, look-up-only,
	// WatchFieldsAuto quirk, subscribe-time request builder). IsSupportedViaAPI and
	// SubscribeManually are derived from ProviderInfo; requiresWatchFieldsAuto and the
	// buildRequestFn payload builder are declared per-provider. Always populated; methods on
	// the zero value are safe to call.
	Subscription SubscriptionConfig

	// Maintenance exposes the periodic-maintenance queries (whether webhook subscriptions
	// require renewal, and the renewal interval). ShouldPerform is derived from ProviderInfo;
	// the renewal interval is declared per-provider via MaintenanceConfig.renewalInterval (no
	// ProviderInfo equivalent today). Always populated; methods on the zero value are safe to
	// call.
	Maintenance MaintenanceConfig

	// PostProcess exposes the post-subscribe third-party setup step (e.g. Salesforce → AWS
	// EventBridge wiring). Only the derived ShouldPerform answer lives here; the setup/teardown
	// functions are Ampersand-infrastructure concerns declared server-side. Always populated;
	// methods on the zero value are safe to call.
	PostProcess PostProcessConfig

	// Verification exposes the webhook-verification and event-receipt concerns (verification
	// params, the webhook-verifier connector, the signature-verification bypass flag, and event
	// casting). All are declared per-provider — there is nothing derived from ProviderInfo
	// here. Always populated; methods on the zero value are safe to call.
	Verification VerificationConfig

	// module, providerInfo, and deps are bound by GetProviderConfig at call time. They flow
	// into each populated subcomponent so subcomponent methods can answer module-aware
	// questions and reach caller-provided capabilities without callers threading the arguments
	// per call.
	module       common.ModuleID
	providerInfo *providers.ProviderInfo
	deps         deps.Dependencies
}

// ProviderConfigRegistry holds a provider's declarative ProviderConfig entries. A provider may
// declare module-specific configs in Modules (keyed by module ID, e.g. "crm", "gmail") and/or a
// DefaultModuleConfig used when no module-specific entry matches the resolved module (or the
// provider has no modules). Mirrors the structure of the connectors-library ProviderInfo.Modules.
type ProviderConfigRegistry struct {
	// DefaultModuleConfig is the config used when the resolved module has no entry in Modules.
	// Nil when the provider declares only module-specific configs.
	DefaultModuleConfig *ProviderConfig

	// Modules holds module-specific configs, keyed by module ID (e.g. ModuleSalesforceCRM "crm",
	// ModuleGoogleGmail "gmail"). Nil/empty when the provider's config is module-agnostic.
	Modules map[common.ModuleID]*ProviderConfig
}

// getConfig resolves the ProviderConfig for the given module: the module-specific entry in
// Modules if one exists, otherwise DefaultModuleConfig. Returns nil when neither is declared.
func (r ProviderConfigRegistry) getConfig(module common.ModuleID) *ProviderConfig {
	if module != "" {
		if cfg, ok := r.Modules[module]; ok {
			return cfg
		}
	}

	return r.DefaultModuleConfig
}

// providerConfigs is the registry of per-provider declarative configs. Each provider maps to a
// ProviderConfigRegistry holding its module-specific configs (Modules) and/or a
// DefaultModuleConfig fallback. GetProviderConfig resolves the module (falling back to
// ProviderInfo.DefaultModule) and calls ProviderConfigRegistry.getConfig.
//
// Aliases (e.g. SalesforceJWT → Salesforce) register the same config pointer under both provider
// keys. providerInfoAliases is used by callers fetching ProviderInfo so that twins missing
// connectors-side SubscribeRequirements resolve correctly.
var providerConfigs = map[providers.Provider]ProviderConfigRegistry{
	providers.Salesforce: {
		Modules: map[common.ModuleID]*ProviderConfig{
			providers.ModuleSalesforceCRM: &salesforceConfig,
		},
	},
	providers.SalesforceJWT: {
		Modules: map[common.ModuleID]*ProviderConfig{
			providers.ModuleSalesforceCRM: &salesforceConfig,
		},
	},
	providers.Zoho: {
		Modules: map[common.ModuleID]*ProviderConfig{
			providers.ModuleZohoCRM:  &zohoConfig,
			providers.ModuleZohoMail: &zohoMailConfig,
		},
	},
	providers.Outreach:  {DefaultModuleConfig: &outreachConfig},
	providers.Salesloft: {DefaultModuleConfig: &salesloftConfig},
	providers.Google: {
		Modules: map[common.ModuleID]*ProviderConfig{
			providers.ModuleGoogleGmail:    &googleConfig,
			providers.ModuleGoogleCalendar: &googleCalendarConfig,
		},
	},
	providers.Hubspot:      {DefaultModuleConfig: &hubspotConfig},
	providers.Gong:         {DefaultModuleConfig: &gongConfig},
	providers.HousecallPro: {DefaultModuleConfig: &housecallproConfig},
	providers.ConnectWise:  {DefaultModuleConfig: &connectWiseConfig},
	providers.AccuLynx:     {DefaultModuleConfig: &acculynxConfig},
	providers.Jobber:       {DefaultModuleConfig: &jobberConfig},
	providers.Slack:        {DefaultModuleConfig: &slackConfig},
	providers.Microsoft:    {DefaultModuleConfig: &microsoftConfig},
	providers.Attio:        {DefaultModuleConfig: &attioConfig},
	providers.Stripe:       {DefaultModuleConfig: &stripeConfig},

	// Subscribe-testing mock providers (see the mocksub package). Inert unless the mock
	// provider is explicitly set up via providers.SetupMock*Provider() in a test.
	providers.MockSalesloft: {DefaultModuleConfig: &mockSalesloftConfig},
}

// ErrProviderConfigNotFound is returned by GetProviderConfig when no entry exists in
// providerConfigs for the resolved (provider, module) pair.
var ErrProviderConfigNotFound = errors.New("provider config not found")

// GetProviderConfig finds the registered ProviderConfig for the given (module, providerInfo)
// pair, applying the resolution order: module → ProviderInfo.DefaultModule → DefaultModuleConfig.
// Returns nil and ErrProviderConfigNotFound if no matching entry exists. On success the returned
// pointer addresses a fresh copy of the registry entry with module/providerInfo/deps bound onto
// it (and onto each subcomponent), so callers receive a fully-bound config without affecting
// registry state.
//
// The module is typically the installation revision's module (empty when the revision declares
// none). deps carries the caller-provided resolver implementations; pass the zero deps.Dependencies
// when none are needed for the concerns being read.
func GetProviderConfig(
	module common.ModuleID,
	providerInfo *providers.ProviderInfo,
	deps deps.Dependencies,
) (*ProviderConfig, error) {
	if providerInfo == nil {
		return nil, fmt.Errorf("%w: nil provider info", ErrProviderConfigNotFound)
	}

	registry, ok := providerConfigs[providerInfo.Name]
	if !ok {
		return nil, fmt.Errorf("%w: provider %q has no entries", ErrProviderConfigNotFound, providerInfo.Name)
	}

	if module == "" {
		module = providerInfo.DefaultModule
	}

	resolved := registry.getConfig(module)
	if resolved == nil {
		return nil, fmt.Errorf("%w: provider %q module %q", ErrProviderConfigNotFound, providerInfo.Name, module)
	}

	// Copy so binding below does not mutate the shared registry entry (which the pointer in
	// providerConfigs addresses, and which aliased providers may share).
	cfg := *resolved

	cfg.module = module
	cfg.providerInfo = providerInfo
	cfg.deps = deps

	cfg.Registration.module = module
	cfg.Registration.providerInfo = providerInfo

	cfg.Subscription.module = module
	cfg.Subscription.providerInfo = providerInfo
	cfg.Subscription.deps = deps

	cfg.Maintenance.module = module
	cfg.Maintenance.providerInfo = providerInfo

	cfg.PostProcess.module = module
	cfg.PostProcess.providerInfo = providerInfo

	cfg.Verification.deps = deps

	return &cfg, nil
}

// ErrRegistrationParamsBuilderNotDeclared is returned by RegistrationConfig.BuildParams when
// the provider's RegistrationConfig has no buildParamsFn declared.
var ErrRegistrationParamsBuilderNotDeclared = errors.New("registration params builder not declared for provider")

// IsProviderLookUpOnly reports whether a provider's webhook subscriptions are managed outside of
// Ampersand. When true, the subscribe workflow persists lookup rows for event routing but skips
// all provider-side create/update/delete calls.
//
// The answer comes from the ProviderInfo: a provider is look-up-only when it supports
// subscriptions (Support.Subscribe = true) but does NOT support programmatic subscription via API
// (SubscribeRequirements.SubscribeByAPI is nil or false).
func IsProviderLookUpOnly(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	if providerInfo == nil {
		return false
	}

	return providerInfoIsLookUpOnly(module, providerInfo)
}

// resolveModuleSubscribeData resolves Support and SubscribeRequirements for a
// (module, providerInfo) pair. Module selection follows the rule: if a non-empty module is
// supplied, that module is the only one consulted (the caller's explicit choice is honored —
// DefaultModule is NOT used as a fallback when the supplied module is missing from
// providerInfo.Modules); otherwise, providerInfo.DefaultModule is used. The selected module is
// looked up in providerInfo.Modules; when it's registered, its Support and SubscribeRequirements
// are used. When the module isn't registered (or no module is selected, or the resolved
// ModuleInfo.SubscribeRequirements is nil), the top-level Support / SubscribeRequirements are
// used as the fallback.
func resolveModuleSubscribeData(
	module common.ModuleID,
	providerInfo *providers.ProviderInfo,
) (providers.Support, *providers.SubscribeRequirements) {
	support := providerInfo.Support

	var requirements *providers.SubscribeRequirements

	if module == "" {
		module = providerInfo.DefaultModule
	}

	if module != "" && providerInfo.Modules != nil {
		if modInfo, ok := (*providerInfo.Modules)[module]; ok {
			support = modInfo.Support
			requirements = modInfo.SubscribeRequirements
		}
	}

	if requirements == nil {
		requirements = providerInfo.SubscribeRequirements
	}

	return support, requirements
}

// providerInfoIsLookUpOnly resolves the look-up-only answer from the ProviderInfo. A
// provider/module is look-up-only when:
//  1. Support.Subscribe is true (the provider supports subscriptions in some form), AND
//  2. SubscribeRequirements.SubscribeByAPI is nil or false (no programmatic API for it — the
//     subscription must be configured manually in the provider UI).
//
// Resolution order for both Support and SubscribeRequirements: module → DefaultModule →
// top-level (see resolveModuleSubscribeData).
func providerInfoIsLookUpOnly(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	support, requirements := resolveModuleSubscribeData(module, providerInfo)

	if !support.Subscribe {
		return false
	}

	return requirements == nil || requirements.SubscribeByAPI == nil || !*requirements.SubscribeByAPI
}

// ProviderRequiresRegistration determines if a provider requires a one-time registration step
// before per-object subscriptions can be created.
//
// The answer comes from the ProviderInfo (per-module SubscribeRequirements.Registration, falling
// back to the provider's DefaultModule, and finally to the top-level
// SubscribeRequirements.Registration).
func ProviderRequiresRegistration(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	// If provider info is nil, we can't determine if registration is required
	if providerInfo == nil {
		return false
	}

	return providerInfoRegistrationRequired(module, providerInfo)
}

// providerInfoRegistrationRequired resolves the SubscribeRequirements.Registration flag from the
// ProviderInfo, using the module → DefaultModule → top-level fallback order of
// resolveModuleSubscribeData. A nil SubscribeRequirements or nil Registration field is treated
// as false.
func providerInfoRegistrationRequired(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	_, requirements := resolveModuleSubscribeData(module, providerInfo)

	if requirements == nil || requirements.Registration == nil {
		return false
	}

	return *requirements.Registration
}

// SubscriptionViaApiSupported reports whether the provider supports programmatic subscription via
// API for the given module.
//
// The answer comes from the ProviderInfo (per-module SubscribeRequirements.SubscribeByAPI,
// falling back to the provider's DefaultModule, and finally to the top-level
// SubscribeRequirements).
func SubscriptionViaApiSupported(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	if providerInfo == nil {
		return false
	}

	return providerInfoSubscribeSupported(module, providerInfo)
}

// providerInfoSubscribeSupported resolves the SubscribeRequirements.SubscribeByAPI flag from the
// ProviderInfo, using the module → DefaultModule → top-level fallback order of
// resolveModuleSubscribeData. A nil SubscribeRequirements or nil SubscribeByAPI field is treated
// as false.
//
// SubscribeByAPI is the right field here (rather than Support.Subscribe): it asks "does the
// provider support programmatic subscription via API?". Support.Subscribe is broader (it can be
// true for providers whose webhooks must be configured manually in the provider UI).
func providerInfoSubscribeSupported(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	_, requirements := resolveModuleSubscribeData(module, providerInfo)

	if requirements == nil || requirements.SubscribeByAPI == nil {
		return false
	}

	return *requirements.SubscribeByAPI
}

// ShouldCreateRegistration determines if registration should be created based on provider info
// and connector. Returns false if the provider doesn't require registration, the connector is
// nil, or the connector doesn't support registration.
func ShouldCreateRegistration(
	module common.ModuleID,
	providerInfo *providers.ProviderInfo,
	connector connectors.SubscribeConnector,
) bool {
	if !ProviderRequiresRegistration(module, providerInfo) {
		return false
	}

	// If connector is nil, we can't create registration
	if connector == nil {
		return false
	}

	_, ok := CastConnector[connectors.RegisterSubscribeConnector](connector)

	return ok
}

// ShouldPostProcess determines if a provider requires a third-party setup step after the
// connector's subscribe call returns (e.g. Salesforce → AWS EventBridge wiring).
//
// The answer comes from the ProviderInfo (per-module SubscribeRequirements.PostProcess, falling
// back to the provider's DefaultModule, and finally to the top-level
// SubscribeRequirements.PostProcess).
func ShouldPostProcess(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	// If provider info is nil, we can't determine if post-processing is required
	if providerInfo == nil {
		return false
	}

	return providerInfoPostProcessRequired(module, providerInfo)
}

// providerInfoPostProcessRequired resolves the SubscribeRequirements.PostProcess flag from the
// ProviderInfo, using the module → DefaultModule → top-level fallback order of
// resolveModuleSubscribeData. A nil SubscribeRequirements or nil PostProcess field is treated as
// false.
func providerInfoPostProcessRequired(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	_, requirements := resolveModuleSubscribeData(module, providerInfo)

	if requirements == nil || requirements.PostProcess == nil {
		return false
	}

	return *requirements.PostProcess
}

// ShouldPerformMaintenance determines if a provider's webhooks require periodic maintenance
// (renewal). For example, Gmail watch subscriptions expire after 7 days and must be re-issued
// before expiry.
//
// The answer comes from the ProviderInfo (per-module SubscribeRequirements.Maintenance, falling
// back to the provider's DefaultModule, and finally to the top-level
// SubscribeRequirements.Maintenance). The renewal interval itself, when maintenance is required,
// is read separately from GetMaintenancePeriod / MaintenanceConfig.Interval.
func ShouldPerformMaintenance(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	if providerInfo == nil {
		return false
	}

	return providerInfoMaintenanceRequired(module, providerInfo)
}

// providerInfoMaintenanceRequired resolves the SubscribeRequirements.Maintenance flag from the
// ProviderInfo, using the module → DefaultModule → top-level fallback order of
// resolveModuleSubscribeData. A nil SubscribeRequirements or nil Maintenance field is treated as
// false.
func providerInfoMaintenanceRequired(module common.ModuleID, providerInfo *providers.ProviderInfo) bool {
	_, requirements := resolveModuleSubscribeData(module, providerInfo)

	if requirements == nil || requirements.Maintenance == nil {
		return false
	}

	return *requirements.Maintenance
}

// connectorWithUnwrap is implemented by connector decorators (e.g. metrics wrappers) that can
// expose the connector they wrap.
type connectorWithUnwrap interface {
	Unwrap() connectors.Connector
}

// CastConnector attempts to cast a connector to a specific type. If the connector is nil, it
// returns a nil value and true. If the cast is successful, it returns the cast connector and
// true. If the cast fails, it attempts to unwrap the connector if it implements the
// connectorWithUnwrap interface and recursively calls CastConnector on the unwrapped connector.
// If the cast is still unsuccessful, it returns a nil value and false.
//
// Mirrors the server's utils.CastConnector so decorated (e.g. metrics-wrapped) connectors passed
// in by the server cast identically.
func CastConnector[ConnectorType connectors.Connector](connector connectors.Connector) (ConnectorType, bool) {
	if connector == nil {
		var empty ConnectorType

		return empty, true
	}

	cast, ok := connector.(ConnectorType)
	if ok {
		return cast, true
	}

	unwrapper, ok := connector.(connectorWithUnwrap)
	if ok {
		return CastConnector[ConnectorType](unwrapper.Unwrap())
	}

	var empty ConnectorType

	return empty, false
}
