package providers

const Smartsheet Provider = "smartsheet"

func init() {
	// Smartsheet Support Configuration
	SetInfo(Smartsheet, ProviderInfo{
		DisplayName: "Smartsheet",
		AuthType:    Oauth2,
		BaseURL:     "https://api.smartsheet.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.smartsheet.com/b/authorize",
			TokenURL:                  "https://api.smartsheet.com/2.0/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722058941/media/smartsheet_1722058939.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722058978/media/smartsheet_1722058967.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722058866/media/smartsheet_1722058865.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722057528/media/smartsheet_1722057527.svg",
			},
		},
	})
}
