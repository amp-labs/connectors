package providers

const Crunchbase Provider = "crunchbase"

func init() {
	// Crunchbase configuration
	SetInfo(Crunchbase, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.crunchbase.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "X-cb-user-key",
			DocsURL:    "https://data.crunchbase.com/docs/getting-started",
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
