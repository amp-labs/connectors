package providers

const Chorus Provider = "chorus"

func init() {
	SetInfo(Chorus, ProviderInfo{
		DisplayName: "Chorus",
		AuthType:    Basic,
		BaseURL:     "https://chorus.ai",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758628755/media/chorus.ai_1758628760.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758628720/media/chorus.ai_1758628721.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758628755/media/chorus.ai_1758628760.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758628720/media/chorus.ai_1758628721.svg",
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
