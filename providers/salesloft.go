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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168309/media/salesloft.com_1722168308.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168309/media/salesloft.com_1722168308.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168309/media/salesloft.com_1722168308.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168309/media/salesloft.com_1722168308.jpg",
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
