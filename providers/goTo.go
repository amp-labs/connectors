package providers

import "github.com/amp-labs/connectors/common"

const (
	GoTo Provider = "goTo"
	// ModuleGoTo covers the api.getgo.com base URL, which serves multiple GoTo
	// products (admin, meetings, webinars, etc). We name it just "goTo" so users
	// don't have to guess which specific product it maps to.
	ModuleGoTo        common.ModuleID = "goTo"
	ModuleGoToConnect common.ModuleID = "goToConnect"
)

func init() {
	SetInfo(GoTo, ProviderInfo{
		DisplayName: "GoTo",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getgo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://authentication.logmeininc.com/oauth/authorize",
			TokenURL:                  "https://authentication.logmeininc.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581742/media/goto.com_1731581740.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581742/media/goto.com_1731581740.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581774/media/goto.com_1731581772.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581774/media/goto.com_1731581772.svg",
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
		PostAuthInfoNeeded: true,
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "accountKey",
				},
			},
		},
		DefaultModule: ModuleGoTo,
		Modules: &Modules{
			ModuleGoTo: {
				BaseURL:     "https://api.getgo.com",
				DisplayName: "GoTo",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
			ModuleGoToConnect: {
				BaseURL:     "https://api.goto.com",
				DisplayName: "GoTo Connect",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
	})
}
