package providers

const OpenAI Provider = "openAI"

func init() {
	SetInfo(OpenAI, ProviderInfo{
		DisplayName: "OpenAI",
		AuthType:    ApiKey,
		BaseURL:     "https://api.openai.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://platform.openai.com/docs/api-reference/api-keys",
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
