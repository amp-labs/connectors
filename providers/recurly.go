package providers

const Recurly Provider = "recurly"

func init() {
	SetInfo(Recurly, ProviderInfo{
		DisplayName: "Recurly",
		AuthType:    Basic,
		BaseURL:     "https://v3.recurly.com",
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722066457/media/recurly_1722066456.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722066434/media/recurly_1722066433.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722066470/media/recurly_1722066469.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722066470/media/recurly_1722066469.svg",
			},
		},
	})
}
