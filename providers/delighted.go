package providers

const Delighted Provider = "delighted"

func init() {
	SetInfo(Delighted, ProviderInfo{
		DisplayName: "Delighted",
		AuthType:    Basic,
		BaseURL:     "https://api.delighted.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733592446/media/Delighted.com_1733592445.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733592497/media/Delighted.com_1733592497.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1733592531/media/Delighted.com_1733592531.png",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1733592557/media/Delighted.com_1733592556.svg",
			},
		},
	})
}
