package providers

const HappyFoxChat Provider = "happyFoxChat"

func init() {
	// happyFox Chat Connector Configuration
	SetInfo(HappyFoxChat, ProviderInfo{
		DisplayName: "HappyFox Chat",
		AuthType:    ApiKey,
		BaseURL:     "https://api.happyfoxchat.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developer.happyfoxchat.com/#authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739528625/media/happyfoxchat.com_1739528625.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739528579/media/happyfoxchat.com_1739528579.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739528625/media/happyfoxchat.com_1739528625.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739528579/media/happyfoxchat.com_1739528579.svg",
			},
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
