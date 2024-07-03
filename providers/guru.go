package providers

const Guru Provider = "guru"

func init() {
	// Guru API Key authentication
	SetInfo(Guru, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://api.getguru.com",
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
	})
}
