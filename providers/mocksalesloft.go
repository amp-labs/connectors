package providers

// MockSalesloft is a subscribe-testing mock provider mimicking Salesloft. Its subscribe config
// reuses Salesloft's SubscriptionEvent types (so event parsing and classification run the real
// code) while the connector serves canned records without HTTP (see the mocksub package).
// Like Mock and MemStore, it is intentionally not registered in init() and must be set up
// manually by calling SetupMockSalesloftProvider().
const MockSalesloft Provider = "mocksalesloft"

// SetupMockSalesloftProvider initializes the MockSalesloft provider configuration. Call this
// explicitly in tests that use the MockSalesloft provider. The Support and
// SubscribeRequirements flags mirror Salesloft's so the subscribe flow takes the same branches.
func SetupMockSalesloftProvider() {
	SetInfo(MockSalesloft, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mocksalesloft.test",
		DisplayName: "Mock Salesloft",
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
