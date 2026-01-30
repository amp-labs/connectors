package providers

const Salesfinity Provider = "salesfinity"

func init() {
	SetInfo(Salesfinity, ProviderInfo{
		DisplayName: "Salesfinity",
		AuthType:    ApiKey,
		BaseURL:     "https://client-api.salesfinity.co",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "x-api-key",
			},
			DocsURL: "https://docs.salesfinity.ai/api-reference/introduction#authentication",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768671066/media/salesfinity.ai_1768671065.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768671062/media/salesfinity.ai_1768671061.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768671066/media/salesfinity.ai_1768671065.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768671062/media/salesfinity.ai_1768671061.png",
			},
		},
	})
}
