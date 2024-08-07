package providers

const AWeber Provider = "aWeber"

func init() {
	// AWeber Configuration
	SetInfo(AWeber, ProviderInfo{
		DisplayName: "AWeber",
		AuthType:    Oauth2,
		BaseURL:     "https://api.aweber.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722164341/media/aWeber_1722164340.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722164323/media/aWeber_1722164322.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722164296/media/aWeber_1722164296.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722164177/media/aWeber_1722164176.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://auth.aweber.com/oauth2/authorize",
			TokenURL:                  "https://auth.aweber.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
