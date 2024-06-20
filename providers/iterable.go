package providers

const Iterable Provider = "iterable"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InHeader,
	}

	if err := apiKeyOpts.MergeApiKeyInHeaderOpts(ApiKeyInHeaderOpts{
		HeaderName: "Api-Key",
		DocsURL:    "https://app.iterable.com/settings/apiKeys",
	}); err != nil {
		panic(err)
	}

	// Iterable API Key authentication
	SetInfo(Iterable, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.iterable.com",
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
