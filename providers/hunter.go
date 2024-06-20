package providers

const Hunter Provider = "hunter"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InQuery,
	}

	if err := apiKeyOpts.MergeApiKeyInQueryParamOpts(ApiKeyInQueryParamOpts{
		QueryParamName: "api_key",
		DocsURL:        "https://hunter.io/api-documentation#authentication",
	}); err != nil {
		panic(err)
	}

	// Hunter Connector Configuration
	SetInfo(Hunter, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.hunter.io/",
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
