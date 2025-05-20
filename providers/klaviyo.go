package providers

import "github.com/amp-labs/connectors/common"

const Klaviyo Provider = "klaviyo"

const (
	// ModuleKlaviyo2024Oct15 is the latest stable version of API as of the date of writing.
	// https://developers.klaviyo.com/en/reference/api_overview
	ModuleKlaviyo2024Oct15 common.ModuleID = "2024-10-15"
)

func init() {
	// Klaviyo configuration
	SetInfo(Klaviyo, ProviderInfo{
		DisplayName: "Klaviyo",
		AuthType:    Oauth2,
		BaseURL:     "https://a.klaviyo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://www.klaviyo.com/oauth/authorize",
			TokenURL:                  "https://a.klaviyo.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
		DefaultModule: ModuleKlaviyo2024Oct15,
		Modules: &Modules{
			ModuleKlaviyo2024Oct15: {
				BaseURL:     "https://a.klaviyo.com",
				DisplayName: "Klaviyo (Version 2024-10-15)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480320/media/klaviyo_1722480318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480320/media/klaviyo_1722480318.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480368/media/klaviyo_1722480367.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480352/media/klaviyo_1722480351.svg",
			},
		},
	})
}
