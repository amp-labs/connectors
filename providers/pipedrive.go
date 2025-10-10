package providers

import "github.com/amp-labs/connectors/common"

const (
	Pipedrive   Provider        = "pipedrive"
	PipedriveV1 common.ModuleID = "v1"
	PipedriveV2 common.ModuleID = "v2"
)

func init() {
	// Pipedrive Configuration
	SetInfo(Pipedrive, ProviderInfo{
		DisplayName: "Pipedrive",
		AuthType:    Oauth2,
		BaseURL:     "https://api.pipedrive.com",
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
		DefaultModule: PipedriveV1,
		Modules: &Modules{
			PipedriveV1: {
				BaseURL:     "https://api.pipedrive.com",
				DisplayName: "Pipedrive v1",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			PipedriveV2: {
				BaseURL:     "https://api.pipedrive.com",
				DisplayName: "Pipedrive v2",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
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
