package providers

import "github.com/amp-labs/connectors/common"

const Mixmax Provider = "mixmax"

func init() {
	SetInfo(Mixmax, ProviderInfo{
		DisplayName: "Mixmax",
		AuthType:    ApiKey,
		BaseURL:     "https://api.mixmax.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-API-Token",
			},
			DocsURL: "https://developer.mixmax.com/reference/getting-started-with-the-api",
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://api.mixmax.com/v1",
				DisplayName: "Mixmax",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722339517/media/mixmax_1722339515.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722339478/media/mixmax_1722339477.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722339517/media/mixmax_1722339515.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722339500/media/mixmax_1722339499.svg",
			},
		},
	})
}
