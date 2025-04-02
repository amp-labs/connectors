package providers

import "github.com/amp-labs/connectors/common"

const ChiliPiper Provider = "chilipiper"

func init() {
	SetInfo(ChiliPiper, ProviderInfo{
		DisplayName: "Chili Piper",
		AuthType:    ApiKey,
		BaseURL:     "https://fire.chilipiper.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
		},
		Modules: &ModuleInfo{
			string(common.ModuleRoot): {
				BaseURL:     "https://fire.chilipiper.com/api/fire-edge/v1/org",
				DisplayName: "Chili Piper",
				Support: ModuleSupport{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
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
