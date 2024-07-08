package providers

const Jotform Provider = "jotform"

func init() {
	// Jotform API Key authentication
	SetInfo(Jotform, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.jotform.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "apiKey",
			},
			DocsURL: "https://api.jotform.com/docs/#authentication",
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
