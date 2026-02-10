package providers

const Granola Provider = "granola"

func init() {
	SetInfo(Granola, ProviderInfo{
		DisplayName: "Granola",
		AuthType:    ApiKey,
		BaseURL:     "https://public-api.granola.ai",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.granola.ai/help-center/sharing/integrations/enterprise-api#how-does-authentication-work",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770680374/media/granola.ai_1770680373.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770680403/media/granola.ai_1770680402.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770680374/media/granola.ai_1770680373.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770680456/media/granola.ai_1770680456.svg",
			},
		},
	})
}
