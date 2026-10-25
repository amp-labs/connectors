package providers

const Circleback Provider = "circleback"

func init() {
	SetInfo(Circleback, ProviderInfo{
		DisplayName: "Circleback",
		AuthType:    ApiKey,
		// The OpenAPI spec's server, https://circleback.ai/api. The API is not versioned.
		BaseURL: "https://circleback.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://circleback.ai/docs/api#get-an-api-key",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				// No separate dark-mode mark was chosen; the regular icon is reused.
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792814059/media/circleback_regular_icon.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792814058/media/circleback_dark_logo.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792814059/media/circleback_regular_icon.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792814060/media/circleback_regular_logo.svg",
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
			Write:     false,
		},
	})
}
