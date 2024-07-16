package providers

const MessageBird Provider = "messageBird"

func init() {
	// MessageBird configuration
	SetInfo(MessageBird, ProviderInfo{
		DisplayName: "MessageBird",
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
