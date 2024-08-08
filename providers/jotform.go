package providers

const Jotform Provider = "jotform"

func init() {
	// Jotform API Key authentication
	SetInfo(Jotform, ProviderInfo{
		DisplayName: "Jotform",
		AuthType:    ApiKey,
		BaseURL:     "https://api.jotform.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "apiKey",
			},
			DocsURL: "https://api.jotform.com/docs/#authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722412121/media/const%20Jotform%20Provider%20%3D%20%22jotform%22_1722412120.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722412311/media/const%20Jotform%20Provider%20%3D%20%22jotform%22_1722412311.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722412121/media/const%20Jotform%20Provider%20%3D%20%22jotform%22_1722412120.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722412311/media/const%20Jotform%20Provider%20%3D%20%22jotform%22_1722412311.svg",
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
	})
}
