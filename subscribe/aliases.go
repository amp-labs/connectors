package subscribe

import (
	"github.com/amp-labs/connectors/providers"
)

// providerInfoAliases maps a provider to another provider whose ProviderInfo it should
// inherit. Used when a provider is a "twin" that reuses another provider's underlying connector
// implementation but has not yet had its own ProviderInfo metadata (e.g. SubscribeRequirements)
// populated.
//
// Lookup direction: key is the provider asking ("alias"), value is the provider whose
// ProviderInfo should be used as the source of truth.
var providerInfoAliases = map[providers.Provider]providers.Provider{
	// SalesforceJWT shares the Salesforce connector implementation and the same modules,
	// but salesforceJWT.go does not yet declare SubscribeRequirements.
	providers.SalesforceJWT: providers.Salesforce,
}

// ResolveProviderInfoAlias returns the ProviderInfo whose data should be used for the given
// provider. If the provider has an entry in providerInfoAliases, this loads the aliased
// provider's ProviderInfo, shallow-copies it, and restores the original .Name so log attribution
// stays accurate. Otherwise the input is returned unchanged.
//
// Callers should apply this once, immediately after fetching ProviderInfo from the providers
// catalog, before passing it into subscribe helpers (SubscriptionViaApiSupported,
// ProviderRequires*, Should*). Those helpers expect an already-resolved ProviderInfo and do not
// re-resolve internally.
func ResolveProviderInfoAlias(providerInfo *providers.ProviderInfo) *providers.ProviderInfo {
	if providerInfo == nil {
		return providerInfo
	}

	aliasTarget, ok := providerInfoAliases[providerInfo.Name]
	if !ok {
		return providerInfo
	}

	aliasedInfo, err := providers.ReadInfo(aliasTarget)
	if err != nil || aliasedInfo == nil {
		return providerInfo
	}

	resolved := *aliasedInfo
	resolved.Name = providerInfo.Name

	return &resolved
}
