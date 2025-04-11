package providers

import "github.com/amp-labs/connectors/common"

const Smartlead Provider = "smartlead"

func init() {
	SetInfo(Smartlead, ProviderInfo{
		DisplayName: "Smartlead",
		AuthType:    ApiKey,
		BaseURL:     "https://server.smartlead.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "api_key",
			},
			DocsURL: "https://api.smartlead.ai/reference/authentication",
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://server.smartlead.ai/api/v1",
				DisplayName: "Smartlead",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/i3juury69prqfujshjly.png",
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
