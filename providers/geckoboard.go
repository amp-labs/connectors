package providers

const Geckoboard Provider = "geckoboard"

func init() {
	SetInfo(Geckoboard, ProviderInfo{
		DisplayName: "Geckoboard",
		AuthType:    Basic,
		BaseURL:     "https://api.geckoboard.com",
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
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071714/media/geckoboard_1722071713.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071714/media/geckoboard_1722071713.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071714/media/geckoboard_1722071713.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071714/media/geckoboard_1722071713.svg",
			},
		},
	})
}
