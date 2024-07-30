package providers

const Bird Provider = "bird"

func init() {
	// Bird configuration
	SetInfo(Bird, ProviderInfo{
		DisplayName: "Bird (MessageBird)",
		AuthType:    ApiKey,
		BaseURL:     "https://api.bird.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "AccessKey ",
			},
			DocsURL: "https://docs.bird.com/api",
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