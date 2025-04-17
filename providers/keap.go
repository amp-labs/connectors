package providers

import "github.com/amp-labs/connectors/common"

const Keap Provider = "keap"

const (
	// ModuleKeapV1 is a grouping of V1 API endpoints.
	// https://developer.keap.com/docs/rest/
	ModuleKeapV1 common.ModuleID = "version1"
	// ModuleKeapV2 is a grouping of V2 API endpoints.
	// https://developer.keap.com/docs/restv2/
	ModuleKeapV2 common.ModuleID = "version2"
)

func init() {
	// Keap configuration
	SetInfo(Keap, ProviderInfo{
		DisplayName: "Keap",
		AuthType:    Oauth2,
		BaseURL:     "https://api.infusionsoft.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.infusionsoft.com/app/oauth/authorize",
			TokenURL:                  "https://api.infusionsoft.com/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Modules: &Modules{
			ModuleKeapV1: {
				BaseURL:     "https://api.infusionsoft.com/v1",
				DisplayName: "Keap Version 1",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
			ModuleKeapV2: {
				BaseURL:     "https://api.infusionsoft.com/v2",
				DisplayName: "Keap Version 2",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724217756/media/Keap_DMI.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479751/media/keap_1722479749.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479775/media/keap_1722479774.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479775/media/keap_1722479774.svg",
			},
		},
	})
}
