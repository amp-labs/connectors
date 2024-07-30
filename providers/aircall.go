package providers

const Aircall Provider = "aircall"

func init() {
	// Aircall Configuration
	SetInfo(Aircall, ProviderInfo{
		DisplayName: "Aircall",
		AuthType:    Oauth2,
		BaseURL:     "https://api.aircall.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dashboard.aircall.io/oauth/authorize",
			TokenURL:                  "https://api.aircall.io/v1/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064159/media/aircall_1722064158.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064141/media/aircall_1722064140.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064183/media/aircall_1722064182.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064183/media/aircall_1722064182.svg",
			},
		},
	})
}
