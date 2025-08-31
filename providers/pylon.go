package providers

const Pylon Provider = "pylon"

func init() {
	// Pylon configuration
	SetInfo(Pylon, ProviderInfo{
		DisplayName: "Pylon",
		AuthType:    ApiKey,
		BaseURL:     "https://api.usepylon.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.usepylon.com/pylon-docs/developer/api",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756274754/media/usepylon.com_1756274753.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756274754/media/usepylon.com_1756274753.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756274730/media/usepylon.com_1756274729.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756274730/media/usepylon.com_1756274729.png",
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
