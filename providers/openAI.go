package providers

const OpenAI Provider = "openAI"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InHeader,
	}

	if err := apiKeyOpts.MergeApiKeyInHeaderOpts(ApiKeyInHeaderOpts{
		HeaderName:  "Authorization",
		ValuePrefix: "Bearer ",
		DocsURL:     "https://platform.openai.com/docs/api-reference/api-keys",
	}); err != nil {
		panic(err)
	}

	SetInfo(OpenAI, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.openai.com",
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
