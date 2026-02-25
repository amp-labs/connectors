package providers

const DevRev Provider = "devrev"

func init() {
	SetInfo(DevRev, ProviderInfo{
		DisplayName: "DevRev",
		AuthType:    Custom,
		BaseURL:     "https://api.devrev.ai",

		CustomOpts: &CustomAuthOpts{
			Headers: []CustomAuthHeader{
				{
					Name:          "Authorization",
					ValueTemplate: "Bearer {{ .token }}",
				},
			},
			Inputs: []CustomAuthInput{
				{
					Name:        "token",
					DisplayName: "Personal Access Token",
					Prompt:      "Personal Access Token",
					DocsURL:     "https://developer.devrev.ai/about/authentication",
				},
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770988610/media/devrev.ai_1770988609.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770988771/media/devrev.ai_1770988771.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770988685/media/devrev.ai_1770988685.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770988737/media/devrev.ai_1770988737.svg",
			},
		},
	})
}
