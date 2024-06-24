package catalog

const OpenAI Provider = "openAI"

func init() {
	SetInfo(OpenAI, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.openai.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader,
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://platform.openai.com/docs/api-reference/api-keys",
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
