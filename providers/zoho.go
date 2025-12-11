package providers

import "github.com/amp-labs/connectors/common"

const (
	Zoho                      Provider        = "zoho"
	ModuleZohoCRM             common.ModuleID = "crm"
	ModuleZohoDesk            common.ModuleID = "desk"
	ModuleZohoProjects        common.ModuleID = "projects"
	ModuleZohoBugTracker      common.ModuleID = "bugtracker"
	ModuleZohoServiceDeskPlus common.ModuleID = "servicedeskplus"
)

// nolint: funlen
func init() {
	// Zoho configuration
	SetInfo(Zoho, ProviderInfo{
		DisplayName: "Zoho",
		AuthType:    Oauth2,
		// E.g. www.zohoapis.com, www.zohoapis.eu, www.zohoapis.in, etc.
		BaseURL:            "https://{{.zoho_api_domain}}",
		PostAuthInfoNeeded: true,
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// NB: This works for all Zoho regions (com, eu, in, cn, au, etc). It will redirect
			// to the appropriate domain based on the user's account. It's ok to hard-code
			// the .com domain here. And since we don't know the user's region in advance,
			// we can't use a templated domain like in BaseURL and TokenURL.
			// See: https://www.zoho.com/crm/developer/docs/api/v8/multi-dc.html
			//
			// Also NB: This won't work for CN region users. They must use accounts.zoho.com.cn
			AuthURL: "https://accounts.zoho.com/oauth/v2/auth",
			// ref: https://www.zoho.com/analytics/api/v2/authentication/generating-code.html
			AuthURLParams: map[string]string{"access_type": "offline"},
			// E.g. accounts.zoho.com, accounts.zoho.eu, accounts.zoho.in, etc.
			TokenURL:                  "https://{{.zoho_token_domain}}/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_domain",
				ScopesField:       "scope",
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
		DefaultModule: ModuleZohoCRM,
		Modules: &Modules{
			ModuleZohoCRM: {
				BaseURL:     "https://{{.zoho_api_domain}}",
				DisplayName: "Zoho CRM",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleZohoDesk: {
				// E.g. www.desk.zoho.com, www.desk.zoho.eu, www.desk.zoho.in, etc.
				BaseURL:     "https://{{.zoho_desk_domain}}",
				DisplayName: "Zoho Desk",
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
			ModuleZohoServiceDeskPlus: {
				BaseURL:     "https://{{.zoho_servicedeskplus_domain}}",
				DisplayName: "Zoho ServiceDeskPlus",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
	})
}
