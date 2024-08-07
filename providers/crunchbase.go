package providers

const Crunchbase Provider = "crunchbase"

func init() {
	// Crunchbase configuration
	SetInfo(Crunchbase, ProviderInfo{
		DisplayName: "Crunchbase",
		AuthType:    ApiKey,
		BaseURL:     "https://api.crunchbase.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-cb-user-key",
			},
			DocsURL: "https://data.crunchbase.com/docs/getting-started",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327098/media/crunchbase_1722327097.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327130/media/crunchbase_1722327129.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327098/media/crunchbase_1722327097.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327157/media/crunchbase_1722327157.svg",
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
