// Package deps declares the resolver seam between the subscribe provider configs and the
// caller (the Ampersand server): narrow interfaces for the capabilities the per-provider
// subscribe functions need but that this library cannot resolve itself (they live behind the
// server's database / customer configuration), plus the request types that cross the boundary.
//
// The server implements these interfaces and passes a Dependencies value to
// subscribe.GetProviderConfig, which binds it onto the returned config so subcomponent methods
// can thread it into the per-provider functions.
package deps

import (
	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/amp-common/webhook"
)

// Dependencies carries the caller-provided resolver implementations.
//
// All fields are optional; a nil resolver means the capability is unavailable. Per-provider
// functions that require a missing resolver return an error (or degrade to their documented
// no-payload behavior — see each use).
type Dependencies struct {
	// Project resolves Ampersand project attributes. Used by the Salesforce subscribe-request
	// builder to read the project's app name for CDC event-flag field naming.
	Project ProjectResolver

	// CDCOptimization resolves the per-(project, group) Salesforce CDC quota-optimization
	// opt-in configuration. Used by the Salesforce subscribe-request builder.
	CDCOptimization CDCOptimizationResolver

	// Subscriptions lists the stored subscription results for an installation. Used by the
	// Attio verification-params builder to recover the webhook signing secret Attio generated
	// at webhook-creation time.
	Subscriptions SubscriptionResultLister
}

// VerificationRequest bundles the inputs to a provider's webhook verification-params builder.
// It replaces the server signature's positional (payload, integration, installation, providerApp)
// parameters with the shared wire types, plus the provider app's OAuth client secret — carried
// separately because the openapi.ProviderApp wire type deliberately excludes secrets.
type VerificationRequest struct {
	// Payload is the inbound webhook delivery being verified.
	Payload *webhook.PubsubPayload

	// Integration is the Ampersand integration the delivery belongs to.
	Integration *openapi.Integration

	// Installation is the Ampersand installation the delivery belongs to.
	Installation *openapi.Installation

	// ProviderApp is the provider app the installation's connection authenticates with.
	ProviderApp *openapi.ProviderApp

	// ProviderAppClientSecret is the provider app's OAuth client secret (not present on the
	// openapi.ProviderApp wire type). Required by providers that sign webhooks with it
	// (Hubspot, Jobber).
	ProviderAppClientSecret string
}
