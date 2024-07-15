package providers

const Apollo = "apollo"

func init() {
	// Apollo API Key authentication
	SetInfo(Apollo, ProviderInfo{
		DisplayName: "Apollo",
		AuthType:    ApiKey,
		BaseURL:     "https://api.apollo.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Api-Key",
			},
			DocsURL: "https://apolloio.github.io/apollo-api-docs/?shell#authentication",
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
