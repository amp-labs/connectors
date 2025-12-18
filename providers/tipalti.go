package providers

const Tipalti Provider = "tipalti"

func init() {
	// Tipalti Configuration
	SetInfo(Tipalti, ProviderInfo{
		DisplayName: "Tipalti",
		AuthType:    Oauth2,
		BaseURL:     "https://api-p.tipalti.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://sso.tipalti.com/connect/authorize/callback",
			TokenURL:                  "https://sso.tipalti.com/connect/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766047240/media/tipalti.com_1766047239.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766047263/media/tipalti.com_1766047263.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766047281/media/tipalti.com_1766047281.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766047296/media/tipalti.com_1766047296.svg",
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
