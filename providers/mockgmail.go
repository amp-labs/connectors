package providers

// MockGmail is a subscribe-testing mock provider mimicking Google Gmail. Its subscribe config
// reuses Gmail's SubscriptionEvent type and caster (so the synthetic republished events the
// receive path consumes parse and classify through the real code) while the connector serves
// canned records without HTTP (see the mocksub package). Like Mock and MemStore, it is
// intentionally not registered in init() and must be set up manually by calling
// SetupMockGmailProvider().
const MockGmail Provider = "mockgmail"

// SetupMockGmailProvider initializes the MockGmail provider configuration. Call this
// explicitly in tests that use the MockGmail provider. The Support and SubscribeRequirements
// flags mirror the Gmail module's (maintenance-renewed watch subscriptions created via API);
// the mock is single-module, so no Modules tree is declared.
func SetupMockGmailProvider() {
	SetInfo(MockGmail, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mockgmail.test",
		DisplayName: "Mock Gmail",
		Support: Support{
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
		SubscribeRequirements: &SubscribeRequirements{
			Maintenance:    new(true),
			SubscribeByAPI: new(true),
		},
	})
}
