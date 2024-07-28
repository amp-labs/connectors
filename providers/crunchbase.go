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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165850/media/crunchbase.com_1722165849.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165850/media/crunchbase.com_1722165849.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165850/media/crunchbase.com_1722165849.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165850/media/crunchbase.com_1722165849.jpg",
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
