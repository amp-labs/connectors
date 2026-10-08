package providers

// MockSalesforce is a subscribe-testing mock provider mimicking Salesforce. Its subscribe
// config reuses Salesforce's SubscriptionEvent types — including the AWS EventBridge envelope
// unwrap and per-recordIds fan-out — so event parsing and classification run the real code,
// while the connector serves canned records without HTTP (see the mocksub package). Like Mock
// and MemStore, it is intentionally not registered in init() and must be set up manually by
// calling SetupMockSalesforceProvider().
const MockSalesforce Provider = "mocksalesforce"

// SetupMockSalesforceProvider initializes the MockSalesforce provider configuration. Call this
// explicitly in tests that use the MockSalesforce provider. SubscribeRequirements is
// deliberately omitted: Salesforce's Registration/PostProcess flags declare Ampersand-side AWS
// EventBridge infrastructure with no mock counterpart, and the webhook receive path under test
// does not read them.
func SetupMockSalesforceProvider() {
	SetInfo(MockSalesforce, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mocksalesforce.test",
		DisplayName: "Mock Salesforce",
		Support: Support{
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
	})
}
