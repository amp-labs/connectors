package providers

// DeepMock is a mock provider with JSON schema validation for testing.
// This provider is intentionally not registered in init() and must be set up manually
// by calling SetupDeepMockProvider(), mirroring the pattern used for the Mock provider.
const DeepMock Provider = "deepmock"

// SetupDeepMockProvider initializes the DeepMock provider configuration.
// You need to call this explicitly if you want to use the DeepMock provider in your tests.
func SetupDeepMockProvider() {
	SetInfo(DeepMock, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://deepmock.memory.store",
		DisplayName: "DeepMock",
		Name:        "deepmock",
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
