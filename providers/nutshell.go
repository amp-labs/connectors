package providers

const Nutshell Provider = "nutshell"

func init() {
	// Nutshell Configuration
	SetInfo(Nutshell, ProviderInfo{
		DisplayName: "Nutshell",
		AuthType:    Basic,
		BaseURL:     "https://app.nutshell.com",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409276/media/const%20Nutshell%20Provider%20%3D%20%22nutshell%22_1722409275.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409318/media/const%20Nutshell%20Provider%20%3D%20%22nutshell%22_1722409317.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409276/media/const%20Nutshell%20Provider%20%3D%20%22nutshell%22_1722409275.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409318/media/const%20Nutshell%20Provider%20%3D%20%22nutshell%22_1722409317.png",
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
