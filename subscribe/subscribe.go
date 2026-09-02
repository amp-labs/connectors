// Package subscribe holds the per-provider declarative subscribe configuration (ProviderConfig)
// and its registry.
//
// A ProviderConfig bundles every subscribe-related concern for one provider — registration,
// subscription request building, maintenance cadence, post-processing requirements, and webhook
// verification — declared as a `var <name>Config = ProviderConfig{...}` literal in
// <provider>.go and registered in providerConfigs (config.go). Callers fetch a bound view via
// GetProviderConfig(module, providerInfo, deps) and read concerns through subcomponent methods
// (e.g. cfg.Subscription.BuildRequest, cfg.Verification.Params).
//
// The contributor model: adding subscribe support for a new provider requires touching just two
// files — the new <provider>.go (declaring the ProviderConfig literal) and config.go (one map
// entry).
//
// This package is a migration of the Ampersand server's shared/subscribe/providers package
// (ENG-4090) so that per-provider subscribe configuration lives alongside the connectors and can
// be contributed without server access. Differences from the server original:
//
//   - Signatures use the shared amp-common wire types (openapi.Installation / Revision /
//     Connection / Integration / ProviderApp) and webhook.PubsubPayload instead of
//     server-internal types.
//   - Module resolution binds a common.ModuleID directly instead of a server revision (only the
//     revision's module was ever read).
//   - Server-owned data (project app names, CDC optimization config, stored subscription
//     results) is reached through the narrow resolver interfaces in deps.Dependencies (the
//     subscribe/deps package), implemented by the server and bound at GetProviderConfig time.
//   - The subscribe-time event-receipt endpoint (webhook URL) is constructed by the caller and
//     passed into BuildRequest.
//   - Post-subscribe third-party setup (e.g. Salesforce → AWS EventBridge wiring) remains
//     server-side; PostProcessConfig here carries only the derived ShouldPerform answer.
//
// File layout mirrors the server package:
//
//   - aliases.go:            ResolveProviderInfoAlias and the providerInfoAliases registry.
//   - config.go:             ProviderConfig + ProviderConfigRegistry + providerConfigs registry +
//     GetProviderConfig + the ProviderInfo derivation helpers.
//   - deps/ (package):       the resolver seam — types.go (Dependencies, VerificationRequest)
//     plus one file per resolver interface (project.go, cdcoptimization.go, subscriptions.go).
//   - events.go:             object-type subscribe-event unwrapping (provider-specific shapes).
//   - maintenance.go:        MaintenanceConfig + maintenancePeriods + GetMaintenancePeriod.
//   - postprocess.go:        PostProcessConfig (derived ShouldPerform only).
//   - registration.go:       RegistrationConfig + methods.
//   - subscription.go:       SubscriptionConfig + methods.
//   - subscriptionevents.go: SubscriptionEventCaster + generic Cast helpers.
//   - verification.go:       VerificationConfig + methods + IsHookdeckGatewayProvider.
//   - <provider>.go:         per-provider <name>Config literal and helpers.
package subscribe

import "errors"

// errInstallationNotFound is a shared sentinel returned by per-provider request builders when an
// installation cannot be found while constructing subscribe params.
var errInstallationNotFound = errors.New("installation not found")

// errNilVerificationRequest is a shared sentinel returned by per-provider verification-params
// builders handed a nil deps.VerificationRequest. Distinct from errInstallationNotFound so
// providers that don't need an installation (integration-scoped webhooks) don't report a
// misleading installation lookup failure.
var errNilVerificationRequest = errors.New("verification request is nil")
