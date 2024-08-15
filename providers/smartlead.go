package providers

const Smartlead Provider = "smartlead"

func init() {
	SetInfo(Smartlead, ProviderInfo{
		DisplayName: "Smartlead AI",
		AuthType:    ApiKey,
		BaseURL:     "https://server.smartlead.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://api.smartlead.ai/reference/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723475823/media/smartlead_1723475823.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723475838/media/smartlead_1723475837.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723475823/media/smartlead_1723475823.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723475838/media/smartlead_1723475837.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
