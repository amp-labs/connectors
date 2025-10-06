package providers

const Linear Provider = "linear"

func init() {
	// Linear configuration
	SetInfo(Linear, ProviderInfo{
		DisplayName: "Linear",
		AuthType:    Oauth2,
		BaseURL:     "https://api.linear.app/graphql",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://linear.app/oauth/authorize",
			TokenURL:                  "https://api.linear.app/oauth/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747229975/media/api.linear.app_1747229974.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747230013/media/api.linear.app_1747230012.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747233254/media/api.linear.app_1747233252.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747230046/media/api.linear.app_1747230045.svg",
			},
		},
	})
}
