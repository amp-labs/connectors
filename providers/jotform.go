package providers

const Jotform Provider = "jotform"

func init() {
	// Jotform API Key authentication
	SetInfo(Jotform, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.jotform.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:           InQuery,
			QueryParamName: "apiKey",
			DocsURL:        "https://api.jotform.com/docs/",
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
