package providers

const Clari Provider = "clari"

func init() {
	SetInfo(Clari, ProviderInfo{
		DisplayName: "Clari",
		AuthType:    ApiKey,
		BaseURL:     "https://api.clari.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "apiKey",
			},
			DocsURL: "https://developer.clari.com/documentation/external_spec",
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
