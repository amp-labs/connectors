package providers

const Anthropic Provider = "anthropic"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InHeader,
	}

	if err := apiKeyOpts.MergeApiKeyInHeaderOpts(ApiKeyInHeaderOpts{
		DocsURL:    "https://docs.anthropic.com/en/api/getting-started#authentication",
		HeaderName: "X-Api-Key",
	}); err != nil {
		panic(err)
	}

	SetInfo(Anthropic, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.anthropic.com",
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
