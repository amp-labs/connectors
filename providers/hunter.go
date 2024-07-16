package providers

const Hunter Provider = "hunter"

func init() {
	// Hunter Connector Configuration
	SetInfo(Hunter, ProviderInfo{
		DisplayName: "Hunter",
		AuthType:    ApiKey,
		BaseURL:     "https://api.hunter.io/",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://hunter.io/api-documentation#authentication",
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
