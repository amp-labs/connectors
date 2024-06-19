package providers

const Guru Provider = "guru"

func init() {
	// Guru API Key authentication
	SetInfo(Guru, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.getguru.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "Api-Key",
			DocsURL:    "https://developer.getguru.com/docs/getting-started",
		},
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
