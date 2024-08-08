package providers

const Guru Provider = "guru"

func init() {
	// Guru API Key authentication
	SetInfo(Guru, ProviderInfo{
		DisplayName: "Guru",
		AuthType:    Basic,
		BaseURL:     "https://api.getguru.com",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410635/media/const%20Guru%20Provider%20%3D%20%22guru%22_1722410634.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410635/media/const%20Guru%20Provider%20%3D%20%22guru%22_1722410634.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410635/media/const%20Guru%20Provider%20%3D%20%22guru%22_1722410634.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410635/media/const%20Guru%20Provider%20%3D%20%22guru%22_1722410634.jpg",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
	})
}
