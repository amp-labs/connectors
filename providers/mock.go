package providers

// Mock is a mock provider that can be used for testing.
// When the mock connector is used, it's added to the catalog
// manually. That is why there's no init() function in this file.
const Mock Provider = "mock"

// SetupMockProvider sets up the mock provider. You need to call
// this explicitly if you want to use the mock provider in your tests.
func SetupMockProvider() {
	SetInfo(Mock, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://mock.web.server",
		DisplayName: "Mock",
		Name:        "mock",
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Delete: true,
				Insert: true,
				Update: true,
				Upsert: true,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
	})
}
