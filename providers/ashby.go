package providers

const Ashby Provider = "ashby"

func init() {
	SetInfo(Ashby, ProviderInfo{
		DisplayName: "ashby",
		AuthType:    Basic,
		BaseURL:     "https://api.ashbyhq.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734352248/media/ashbyhq.com_1734352247.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734352288/media/ashbyhq.com_1734352288.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734352248/media/ashbyhq.com_1734352247.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734352330/media/ashbyhq.com_1734352330.svg",
			},
		},
	})
}
