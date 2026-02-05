package providers

// MemStore is a mock provider with JSON schema validation for testing.
// This provider is intentionally not registered in init() and must be set up manually
// by calling SetupMemStoreProvider(), mirroring the pattern used for the Mock provider.
const MemStore Provider = "memstore"

// SetupMemStoreProvider initializes the MemStore provider configuration.
// You need to call this explicitly if you want to use the MemStore provider in your tests.
func SetupMemStoreProvider() {
	SetInfo(MemStore, ProviderInfo{
		AuthType:    None,
		BaseURL:     "http://memory.store",
		DisplayName: "Memory Store",
		Name:        "memstore",
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
