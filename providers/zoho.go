package providers

import "github.com/amp-labs/connectors/common"

const (
	Zoho Provider = "zoho"
)

const (
	ModuleZoho           common.ModuleID = "zoho"
	ModuleZohoProjects   common.ModuleID = "projects"
	ModuleZohoBugTracker common.ModuleID = "bugtracker"
)

//nolint:funlen
func init() {
	// Zoho configuration
	SetInfo(Zoho, ProviderInfo{
		DisplayName: "Zoho",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zohoapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			AuthURL:   "https://accounts.zoho.com/oauth/v2/auth",
			// ref: https://www.zoho.com/analytics/api/v2/authentication/generating-code.html
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_domain",
				ScopesField:       "scope",
			},
		},
		DefaultModule: ModuleZoho,
		Modules: &Modules{
			ModuleZoho: {
				BaseURL:     "https://www.zohoapis.com",
				DisplayName: "Zoho",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleZohoProjects: {
				BaseURL:     "https://projectsapi.zoho.com",
				DisplayName: "Zoho Projects",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},

			ModuleZohoBugTracker: {
				BaseURL:     "https://bugtracker.zoho.com",
				DisplayName: "Zoho BugTracker",
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724224295/media/lk7ohfgtmzys1sl919c8.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471872/media/zoho_1722471871.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471890/media/zoho_1722471890.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471890/media/zoho_1722471890.svg",
			},
		},
	})
}
