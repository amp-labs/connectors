package providers

const Github Provider = "github"

func init() {
	SetInfo(Github, ProviderInfo{
		DisplayName: "GitHub",
		AuthType:    Oauth2,
		BaseURL:     "https://api.github.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://github.com/login/oauth/authorize",
			TokenURL:                  "https://github.com/login/oauth/access_token",
			ExplicitScopesRequired:    true,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722449305/media/github_1722449304.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722449225/media/github_1722449224.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722449256/media/github_1722449255.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722449198/media/github_1722449197.png",
			},
		},
	})
}
