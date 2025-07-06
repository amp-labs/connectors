package providers

const (
	Zoho           Provider = "zoho"
	ZohoProjects   Provider = "zohoProjects"
	ZohoBugTracker Provider = "zohoBugTracker"
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

	SetInfo(ZohoProjects, ProviderInfo{
		DisplayName: "Zoho Projects",
		AuthType:    Oauth2,
		BaseURL:     "https://projectsapi.zoho.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739333796/oda23pivlucm71ef7vu_bbv5sl.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739333472/projects_lw3z3y.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739333796/oda23pivlucm71ef7vu_bbv5sl.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739333472/projects_lw3z3y.svg",
			},
		},
	})

	SetInfo(ZohoBugTracker, ProviderInfo{
		DisplayName: "Zoho BugTracker",
		AuthType:    Oauth2,
		BaseURL:     "https://bugtracker.zoho.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739342042/r25of246tym71jbt61_kuiugm.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739342060/bugtracker_qtjkcz.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739342042/r25of246tym71jbt61_kuiugm.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739342060/bugtracker_qtjkcz.svg",
			},
		},
	})
}
