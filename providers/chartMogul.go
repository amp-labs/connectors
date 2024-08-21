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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724223755/media/dofc2atuowphyzyh3x4l.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/yrna2ica74nfjgxmkie0.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071196/media/chartMogul_1722071194.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071151/media/chartMogul_1722071150.svg",
			},
		},
	})
}
