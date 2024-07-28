package providers

const Anthropic Provider = "anthropic"

func init() {
	SetInfo(Anthropic, ProviderInfo{
		DisplayName: "Anthropic",
		AuthType:    ApiKey,
		BaseURL:     "https://api.anthropic.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "x-api-key",
			},
			DocsURL: "https://docs.anthropic.com/en/api/getting-started#authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184511/media/anthropic_1722184511.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184511/media/anthropic_1722184511.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184511/media/anthropic_1722184511.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184511/media/anthropic_1722184511.jpg",
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
