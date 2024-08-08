package providers

const OpenAI Provider = "openAI"

func init() {
	SetInfo(OpenAI, ProviderInfo{
		DisplayName: "OpenAI",
		AuthType:    ApiKey,
		BaseURL:     "https://api.openai.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://platform.openai.com/docs/api-reference/api-keys",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348143/media/openAI_1722348141.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348197/media/openAI_1722348196.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348143/media/openAI_1722348141.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348211/media/openAI_1722348211.svg",
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
