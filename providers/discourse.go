package providers

const Discourse Provider = "discourse"

func init() {
	SetInfo(Discourse, ProviderInfo{
		DisplayName: "Discourse",
		AuthType:    Custom,
		BaseURL:     "https://{{.workspace}}",
		CustomOpts: &CustomAuthOpts{
			Headers: []CustomAuthHeader{
				{
					Name:          "Api-Key",
					ValueTemplate: "{{ .apiKey }}",
				},
				{
					Name:          "Api-Username",
					ValueTemplate: "{{ .apiUsername }}",
				},
			},
			Inputs: []CustomAuthInput{
				{
					Name:        "apiKey",
					DisplayName: "API Key",
				},
				{
					Name:        "apiUsername",
					DisplayName: "API Username",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734557159/media/discourse.org_1734557159.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734557186/media/discourse.org_1734557186.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1734557116/media/discourse.org_1734557115.jpg",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1734557138/media/discourse.org_1734557138.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Your domain",
				},
			},
		},
	})
}
