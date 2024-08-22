package providers

const WordPress Provider = "wordPress"

func init() {
	// WordPress Support configuration
	SetInfo(WordPress, ProviderInfo{
		DisplayName: "WordPress",
		AuthType:    Oauth2,
		BaseURL:     "https://public-api.wordpress.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225616/media/jwc5dcjfheo0vpr8e1ga.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225616/media/jwc5dcjfheo0vpr8e1ga.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346246/media/const%20WordPress%20Provider%20%3D%20%22wordPress%22_1722346246.svg,
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346154/media/const%20WordPress%20Provider%20%3D%20%22wordPress%22_1722346154.svg",
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
	})
}
