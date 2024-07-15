package providers

const Coda Provider = "coda"

func init() {
	// Coda Configuration
	SetInfo(Coda, ProviderInfo{
		DisplayName: "Coda",
		AuthType:    ApiKey,
		BaseURL:     "https://coda.io/apis",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://coda.io/developers/apis/v1#section/Introduction",
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
