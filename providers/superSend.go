package providers

const SuperSend Provider = "superSend"

func init() {
	// SuperSend API Key authentication
	// API documentation: https://docs.supersend.io
	SetInfo(SuperSend, ProviderInfo{
		DisplayName: "Super Send",
		AuthType:    ApiKey,
		BaseURL:     "https://api.supersend.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.supersend.io/docs/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768601185/media/supersend.io_1768601185.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768601134/media/supersend.io_1768601132.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768601185/media/supersend.io_1768601185.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768601134/media/supersend.io_1768601132.svg",
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
