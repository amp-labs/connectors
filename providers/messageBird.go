package providers

const MessageBird Provider = "messageBird"

func init() {
	// MessageBird configuration
	SetInfo(MessageBird, ProviderInfo{
		DisplayName: "MessageBird",
		AuthType:    ApiKey,
		BaseURL:     "https://api.bird.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166753/media/messagebird.com_1722166752.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166753/media/messagebird.com_1722166752.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166753/media/messagebird.com_1722166752.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166753/media/messagebird.com_1722166752.jpg",
			},
		},
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
