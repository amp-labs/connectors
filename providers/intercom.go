package providers

const Intercom Provider = "intercom"

func init() {
	// Intercom configuration
	SetInfo(Intercom, ProviderInfo{
		DisplayName: "Intercom",
		AuthType:    Oauth2,
		BaseURL:     "https://api.intercom.io",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166109/media/intercom.com_1722166108.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327671/media/intercom_1722327670.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166109/media/intercom.com_1722166108.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327671/media/intercom_1722327670.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.intercom.com/oauth",
			TokenURL:                  "https://api.intercom.io/auth/eagle/token",
			ExplicitScopesRequired:    false,
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
