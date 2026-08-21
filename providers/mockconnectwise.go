package providers

// MockConnectWise is a subscribe-testing mock provider mimicking ConnectWise. Its subscribe
// config reuses ConnectWise's SubscriptionEvent types — including the inline-record Entity
// extraction (SubscriptionEventWithRecord) — so event parsing and classification run the real
// code, while the connector serves canned records without HTTP (see the mocksub package). Like
// Mock and MemStore, it is intentionally not registered in init() and must be set up manually
// by calling SetupMockConnectWiseProvider().
const MockConnectWise Provider = "mockconnectwise"

// SetupMockConnectWiseProvider initializes the MockConnectWise provider configuration. Call
// this explicitly in tests that use the MockConnectWise provider. The Support and
// SubscribeRequirements flags mirror ConnectWise's so the subscribe flow takes the same
// branches; the region metadata input is omitted (the mock talks to no HTTP API).
func SetupMockConnectWiseProvider() {
	SetInfo(MockConnectWise, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mockconnectwise.test",
		DisplayName: "Mock ConnectWise",
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
