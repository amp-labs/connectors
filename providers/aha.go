package providers

import "github.com/amp-labs/connectors/common"

const Aha Provider = "aha"

func init() {
	// Aha Configuration
	SetInfo(Aha, ProviderInfo{
		DisplayName: "Aha",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.aha.io/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.aha.io/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.aha.io/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
		},
		Modules: &ModuleInfo{
			string(common.ModuleRoot): {
				BaseURL:     "https://{{.workspace}}.aha.io/api/v1",
				DisplayName: "Aha",
				Support: ModuleSupport{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347563/media/aha_1722347563.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347588/media/aha_1722347588.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347563/media/aha_1722347563.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347605/media/aha_1722347605.svg",
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
