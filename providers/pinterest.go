package providers

const Pinterest Provider = "pinterest"

func init() {
	// Pinterest Configuration
	SetInfo(Pinterest, ProviderInfo{
		DisplayName: "Pinterest",
		AuthType:    Oauth2,
		BaseURL:     "https://api.pinterest.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.pinterest.com/oauth",
			TokenURL:                  "https://api.pinterest.com/v5/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405637/media/const%20Pinterest%20Provider%20%3D%20%22pinterest%22_1722405635.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405701/media/const%20Pinterest%20Provider%20%3D%20%22pinterest%22_1722405701.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405637/media/const%20Pinterest%20Provider%20%3D%20%22pinterest%22_1722405635.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405701/media/const%20Pinterest%20Provider%20%3D%20%22pinterest%22_1722405701.svg",
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
