package providers

const Instantly Provider = "instantly"

func init() {
	SetInfo(Instantly, ProviderInfo{
		DisplayName: "Instantly",
		AuthType:    ApiKey,
		BaseURL:     "https://api.instantly.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://developer.instantly.ai/introduction",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
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
