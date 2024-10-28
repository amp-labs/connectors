package providers

const Kit Provider = "kit"

func init() {
	// Kit configuration
	SetInfo(Kit, ProviderInfo{
		DisplayName: "Kit",
		AuthType:    Oauth2,
		BaseURL:     "https://api.kit.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.kit.com/oauth/authorize",
			TokenURL:                  "https://app.kit.com/oauth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://kit.com/favicon.ico",
				LogoURL: "https://media.kit.com/images/logos/kit-logo-warm-white.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://kit.com/favicon.ico",
				LogoURL: "https://media.kit.com/images/logos/kit-logo-soft-black.svg",
			},
		},
	})
}
