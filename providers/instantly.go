package providers

import "github.com/amp-labs/connectors/common"

const (
	Instantly   Provider = "instantly"
	InstantlyAI Provider = "instantlyAI"
)

//nolint:funlen
func init() {
	// Instantly v1 configuration
	SetInfo(Instantly, ProviderInfo{
		DisplayName: "Instantly (Legacy V1)",
		AuthType:    ApiKey,
		BaseURL:     "https://api.instantly.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://developer.instantly.ai/introduction",
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://api.instantly.ai/api/v1",
				DisplayName: "Instantly (Legacy V1)",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
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

	// Instantly v2 configuration
	SetInfo(InstantlyAI, ProviderInfo{
		DisplayName: "Instantly",
		AuthType:    ApiKey,
		BaseURL:     "https://api.instantly.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developer.instantly.ai/getting-started/getting-started#generate-a-new-api-key",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645909/media/instantly_1723645909.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723645924/media/instantly_1723645924.svg",
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
