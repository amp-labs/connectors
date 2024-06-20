package providers

const Guru Provider = "guru"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InHeader,
	}

	if err := apiKeyOpts.MergeApiKeyInHeaderOpts(ApiKeyInHeaderOpts{
		HeaderName: "Api-Key",
		DocsURL:    "https://developer.getguru.com/docs/getting-started",
	}); err != nil {
		panic(err)
	}

	// Guru API Key authentication
	SetInfo(Guru, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.getguru.com",
		ApiKeyOpts: apiKeyOpts,
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
