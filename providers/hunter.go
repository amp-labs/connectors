package providers

const Hunter Provider = "hunter"

func init() {
	// Hunter Connector Configuration
	SetInfo(Hunter, ProviderInfo{
		DisplayName: "Hunter",
		AuthType:    ApiKey,
		BaseURL:     "https://api.hunter.io/",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://hunter.io/api-documentation#authentication",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456821/media/hunter_1722456820.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456804/media/hunter_1722456803.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456821/media/hunter_1722456820.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456762/media/hunter_1722456761.svg",
			},
		},
	})
}
