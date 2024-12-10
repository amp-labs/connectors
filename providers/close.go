package providers

const Close Provider = "close"

func init() {
	// Close Configuration
	SetInfo(Close, ProviderInfo{
		DisplayName: "Close",
		AuthType:    Oauth2,
		BaseURL:     "https://api.close.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.close.com/oauth2/authorize",
			TokenURL:                  "https://api.close.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "user_id",
				WorkspaceRefField: "organization_id",
				ScopesField:       "scope",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722513593/media/const%20Close%20Provider%20%3D%20%22close%22_1722513594.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722513669/media/const%20Close%20Provider%20%3D%20%22close%22_1722513670.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722513593/media/const%20Close%20Provider%20%3D%20%22close%22_1722513594.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722513650/media/const%20Close%20Provider%20%3D%20%22close%22_1722513652.svg",
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
	})
}
