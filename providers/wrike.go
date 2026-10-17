package providers

const (
	Wrike Provider = "wrike"
	// WrikeEU serves accounts hosted in Wrike's EU data centre (app-eu.wrike.com).
	WrikeEU Provider = "wrikeEU"
	// WrikeUS2 serves accounts hosted in Wrike's second US data centre (app-us2.wrike.com).
	WrikeUS2 Provider = "wrikeUS2"
)

//nolint:funlen
func init() {
	// Wrike configuration
	SetInfo(Wrike, ProviderInfo{
		DisplayName: "Wrike",
		AuthType:    Oauth2,
		BaseURL:     "https://www.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
			ExplicitScopesRequired:    true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471561/media/wrike_1722471561.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471516/media/wrike_1722471514.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471586/media/wrike_1722471585.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471537/media/wrike_1722471536.svg",
			},
		},
	})

	// Wrike stores each account in one data centre and its API answers only on
	// that host (other hosts return 401). The OAuth token response names it in
	// `host`. OAuth endpoints are shared: the code exchange and refresh work on
	// any host (refresh verified live on app-eu for a www.wrike.com account).
	// WrikeEU configuration
	SetInfo(WrikeEU, ProviderInfo{
		DisplayName: "Wrike (EU)",
		AuthType:    Oauth2,
		BaseURL:     "https://app-eu.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
			ExplicitScopesRequired:    true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471561/media/wrike_1722471561.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471516/media/wrike_1722471514.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471586/media/wrike_1722471585.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471537/media/wrike_1722471536.svg",
			},
		},
	})

	// WrikeUS2 configuration
	SetInfo(WrikeUS2, ProviderInfo{
		DisplayName: "Wrike (US Domain)",
		AuthType:    Oauth2,
		BaseURL:     "https://app-us2.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
			ExplicitScopesRequired:    true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471561/media/wrike_1722471561.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471516/media/wrike_1722471514.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471586/media/wrike_1722471585.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471537/media/wrike_1722471536.svg",
			},
		},
	})
}
