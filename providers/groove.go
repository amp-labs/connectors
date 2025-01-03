package providers

const Groove Provider = "groove"

func init() {
	// Groove configuration
	SetInfo(Groove, ProviderInfo{
		DisplayName: "Groove",
		AuthType:    Oauth2,
		BaseURL:     "https://api.groovehq.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://www.groovehq.com/images/v3/brand/groove-lettermark-light.png",
				LogoURL: "https://www.groovehq.com/images/v3/brand/groove-wordmark.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://www.groovehq.com/images/v3/brand/groove-lettermark.png",
				LogoURL: "https://www.groovehq.com/images/v3/brand/official-groove-logo.png",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.groovehq.com/oauth/authorize",
			TokenURL:                  "https://api.groovehq.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
	})
}
