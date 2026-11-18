package objects

import "github.com/amp-labs/connectors/providers"

// genericProviders are the providers that operate on whatever object the credential can
// reach, rather than on a fixed list. There is nothing to generate a complete object list
// from for these -- the list depends on the customer's instance, not on the provider.
//
//nolint:gochecknoglobals
var genericProviders = map[string]bool{
	providers.Salesforce:    true,
	providers.SalesforceJWT: true,
	providers.Hubspot:       true,
	providers.DynamicsCRM:   true,
	providers.Close:         true,
}

// declaredObjects enumerates objects known to work on generic providers.
//
// These are a convenience, not a boundary: a generic provider can reach objects that are
// not listed here, which is what ProviderObjects.Generic exists to say. What they buy is an
// answer to "can this provider do activity objects at all" that does not require standing
// up a connection first.
//
// The activity objects are here because they are what actually gets asked about -- calls,
// emails, meetings and messages are the objects an engagement platform syncs, and they were
// the least documented. Each list is asserted against the connector's own knowledge where
// the connector has any, so the two cannot drift.
//
//nolint:gochecknoglobals
var declaredObjects = map[string]map[string][]string{
	providers.Salesforce: {
		// Salesforce activity objects. Task and Event are the classic pair; EmailMessage
		// and VoiceCall are the ones engagement platforms need and the ones nothing in
		// this repo previously mentioned.
		"": {"Task", "Event", "EmailMessage", "VoiceCall"},
	},
	providers.SalesforceJWT: {
		"": {"Task", "Event", "EmailMessage", "VoiceCall"},
	},
	providers.Hubspot: {
		// Mirrors the activity subset of hubspot.KnownObjectTypes, which is itself keyed
		// to HubSpot's published object type IDs. `communications` is the object behind
		// SMS, WhatsApp and LinkedIn messages, and is the one most often missed.
		"crm": {"calls", "communications", "emails", "meetings", "notes", "tasks"},
	},
	providers.DynamicsCRM: {
		// Dynamics entity set names, which are what the connector addresses.
		"": {"tasks", "appointments", "phonecalls", "emails"},
	},
}

// IsGeneric reports whether a provider operates on whatever object the credential can reach
// rather than on a fixed list.
func IsGeneric(provider string) bool {
	return genericProviders[provider]
}

// DeclaredObjects returns the hand-declared objects for a provider, keyed by module.
func DeclaredObjects(provider string) map[string][]string {
	return declaredObjects[provider]
}
