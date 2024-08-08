package providers

const Salesloft Provider = "salesloft"

func init() {
	// Salesloft configuration
	SetInfo(Salesloft, ProviderInfo{
		DisplayName: "Salesloft",
		AuthType:    Oauth2,
		BaseURL:     "https://api.salesloft.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330218/media/salesloft_1722330216.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330241/media/salesloft_1722330240.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330218/media/salesloft_1722330216.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330274/media/salesloft_1722330273.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:                  "https://accounts.salesloft.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
