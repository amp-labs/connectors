package providers

const Crunchbase Provider = "crunchbase"

func init() {
	// Crunchbase configuration
	SetInfo(Crunchbase, ProviderInfo{
		DisplayName: "Crunchbase",
		AuthType:    ApiKey,
		BaseURL:     "https://api.crunchbase.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-cb-user-key",
			},
			DocsURL: "https://data.crunchbase.com/docs/getting-started",
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
