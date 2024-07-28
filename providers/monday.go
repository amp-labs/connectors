package providers

const Monday Provider = "monday"

func init() {
	// Monday Configuration
	SetInfo(Monday, ProviderInfo{
		DisplayName: "Monday",
		AuthType:    Oauth2,
		BaseURL:     "https://api.monday.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.monday.com/oauth2/authorize",
			TokenURL:                  "https://auth.monday.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722182530/media/monday_1722182528.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722182530/media/monday_1722182528.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722182530/media/monday_1722182528.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722182530/media/monday_1722182528.jpg",
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
