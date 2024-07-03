package providers

const Dixa Provider = "dixa"

func init() {
	SetInfo(Dixa, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://dev.dixa.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://docs.dixa.io/docs/api-standards-rules/#authentication",
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
