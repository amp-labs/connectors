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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328501/media/bird_1722328500.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328456/media/bird_1722328455.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328535/media/bird_1722328534.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328476/media/bird_1722328475.png",
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
