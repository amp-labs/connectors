package providers

const Fireflies Provider = "fireflies"

func init() {
	// Fireflies Configuration
	SetInfo(Fireflies, ProviderInfo{
		DisplayName: "Fireflies",
		AuthType:    ApiKey,
		BaseURL:     "https://api.fireflies.ai/graphql",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: "header",
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.fireflies.ai/fundamentals/authorization#acquiring-a-token",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1745419046/media/fireflies.ai_1745419046.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1745419015/media/fireflies.ai_1745419014.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1745419046/media/fireflies.ai_1745419046.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1745419073/media/fireflies.ai_1745419073.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
