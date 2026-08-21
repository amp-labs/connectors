package providers

// MockHubspot is a subscribe-testing mock provider mimicking HubSpot. Its subscribe config
// reuses HubSpot's SubscriptionEvent types and caster (so event parsing and classification run
// the real code) while the connector serves canned records without HTTP (see the mocksub
// package). Like Mock and MemStore, it is intentionally not registered in init() and must be
// set up manually by calling SetupMockHubspotProvider().
const MockHubspot Provider = "mockhubspot"

// SetupMockHubspotProvider initializes the MockHubspot provider configuration. Call this
// explicitly in tests that use the MockHubspot provider. Like HubSpot, it declares no
// SubscribeRequirements: HubSpot subscriptions are managed at the provider-app level
// (look-up-only), not created via API.
func SetupMockHubspotProvider() {
	SetInfo(MockHubspot, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mockhubspot.test",
		DisplayName: "Mock HubSpot",
		Support: Support{
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
	})
}
