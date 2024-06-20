package providers

const Apollo = "apollo"

func init() {
	// Apollo API Key authentication
	SetInfo(Apollo, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.apollo.io",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "Api-Key",
			DocsURL:    "https://app.apollo.io/#/settings/integrations/api",
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
