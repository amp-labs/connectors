package providers

// MockAttio is a subscribe-testing mock provider mimicking Attio. Its subscribe config reuses
// Attio's SubscriptionEvent types (so event parsing and classification run the real code)
// while the connector serves canned records — and resolves record.* events' id.object_id via a
// seeded index instead of the Attio API — without HTTP (see the mocksub package). Like Mock
// and MemStore, it is intentionally not registered in init() and must be set up manually by
// calling SetupMockAttioProvider().
const MockAttio Provider = "mockattio"

// SetupMockAttioProvider initializes the MockAttio provider configuration. Call this
// explicitly in tests that use the MockAttio provider. The Support and SubscribeRequirements
// flags mirror Attio's so the subscribe flow takes the same branches.
func SetupMockAttioProvider() {
	SetInfo(MockAttio, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mockattio.test",
		DisplayName: "Mock Attio",
		Support: Support{
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
		SubscribeRequirements: &SubscribeRequirements{
			SubscribeByAPI: new(true),
		},
	})
}
