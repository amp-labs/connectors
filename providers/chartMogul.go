package providers

const ChartMogul Provider = "chartMogul"

func init() {
	SetInfo(ChartMogul, ProviderInfo{
		DisplayName: "ChartMogul",
		AuthType:    Basic,
		BaseURL:     "https://api.chartmogul.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071123/media/chartMogul_1722071122.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071123/media/chartMogul_1722071122.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071196/media/chartMogul_1722071194.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071151/media/chartMogul_1722071150.svg",
			},
		},
	})
}
