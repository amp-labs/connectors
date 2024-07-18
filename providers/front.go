package providers

const Front Provider = "front"

func init() {
	SetInfo(Front, ProviderInfo{
		DisplayName: "Front",
		AuthType:    ApiKey,
		BaseURL:     "https://api2.frontapp.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://dev.frontapp.com/docs/create-and-revoke-api-tokens",
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
