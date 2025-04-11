package providers

import "github.com/amp-labs/connectors/common"

const Pipedrive Provider = "pipedrive"

func init() {
	// Pipedrive Configuration
	SetInfo(Pipedrive, ProviderInfo{
		DisplayName: "Pipedrive",
		AuthType:    Oauth2,
		BaseURL:     "https://api.pipedrive.com",
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://api.pipedrive.com/v1",
				DisplayName: "Pipedrive",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://oauth.pipedrive.com/oauth/authorize",
			TokenURL:                  "https://oauth.pipedrive.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470001/media/pipedrive_1722470000.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469920/media/pipedrive_1722469919.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469947/media/pipedrive_1722469947.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469899/media/pipedrive_1722469898.svg",
			},
		},
	})
}
