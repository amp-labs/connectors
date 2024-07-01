package catalog

const Anthropic Provider = "anthropic"

func init() {
	SetInfo(Anthropic, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.anthropic.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "x-api-key",
			DocsURL:    "https://docs.anthropic.com/en/api/getting-started#authentication",
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
