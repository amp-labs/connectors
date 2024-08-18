package providers

const Gladly = "gladly"

func init() {
	SetInfo(Gladly, ProviderInfo{
		DisplayName: "Gladly",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}/api",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723973960/media/gladly_1723973958.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723974024/media/gladly_1723974023.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723973960/media/gladly_1723973958.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723974024/media/gladly_1723974023.jpg",
			},
		},
	})
}
