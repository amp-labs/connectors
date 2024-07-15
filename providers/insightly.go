package providers

const Insightly = "insightly"

func init() {
	// Insightly API Key authentication
	SetInfo(Insightly, ProviderInfo{
		DisplayName: "Insightly",
		AuthType:    Basic,
		BaseURL:     "https://api.insightly.com",
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
	})
}
