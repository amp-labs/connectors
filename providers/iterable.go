package providers

const Iterable Provider = "iterable"

func init() {
	// Iterable API Key authentication
	SetInfo(Iterable, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.iterable.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Api-Key",
			},
			DocsURL: "https://app.iterable.com/settings/apiKeys",
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
