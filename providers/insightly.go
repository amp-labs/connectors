package providers

const Insightly = "insightly"

func init() {
	// Insightly API Key authentication
	SetInfo(Insightly, ProviderInfo{
		DisplayName: "Insightly",
		AuthType:    Basic,
		BaseURL:     "https://api.insightly.com",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411056/media/const%20Insightly%20%3D%20%22insightly%22_1722411055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411001/media/const%20Insightly%20%3D%20%22insightly%22_1722411000.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411056/media/const%20Insightly%20%3D%20%22insightly%22_1722411055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411001/media/const%20Insightly%20%3D%20%22insightly%22_1722411000.svg",
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
