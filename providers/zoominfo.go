package providers

const Zoominfo Provider = "zoominfo"

func init() {
	// Zoominfo configuration
	SetInfo(Zoominfo, ProviderInfo{
		DisplayName: "Zoominfo",
		AuthType:    Basic,
		BaseURL:     "https://api.zoominfo.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621074/media/zoominfo.com_1758621078.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621183/media/zoominfo.com_1758621188.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621074/media/zoominfo.com_1758621078.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621167/media/zoominfo.com_1758621172.svg",
			},
		},
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
	})
}
