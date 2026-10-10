package providers

const Grain Provider = "grain"

func init() {
	SetInfo(Grain, ProviderInfo{
		DisplayName: "Grain",
		AuthType:    Oauth2,
		BaseURL:     "https://api.grain.com/_/public-api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://grain.com/_/public-api/oauth2/authorize",
			TokenURL:                  "https://api.grain.com/_/public-api/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			ScopeQueryParam:           "scopes",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1784041266/media/grain.com_1784041265.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1784041233/media/grain.com_1784041231.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1784041266/media/grain.com_1784041265.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1784041233/media/grain.com_1784041231.svg",
			},
		},
	})
}
