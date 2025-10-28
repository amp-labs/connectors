package providers

const (
	Ramp     Provider = "ramp"
	RampDemo Provider = "rampDemo"
)

func init() { //nolint:funlen
	SetInfo(Ramp, ProviderInfo{
		DisplayName: "Ramp",
		AuthType:    Oauth2,
		BaseURL:     "https://api.ramp.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.ramp.com/v1/authorize",
			TokenURL:                  "https://api.ramp.com/developer/v1/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877594/media/ramp.com_1758877593.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877218/media/ramp.com_1758877216.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877594/media/ramp.com_1758877593.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877294/media/ramp.com_1758877293.svg",
			},
		},
	})

	SetInfo(RampDemo, ProviderInfo{
		DisplayName: "Ramp Demo",
		AuthType:    Oauth2,
		BaseURL:     "https://demo-api.ramp.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://demo.ramp.com/v1/authorize",
			TokenURL:                  "https://demo-api.ramp.com/developer/v1/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877594/media/ramp.com_1758877593.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877218/media/ramp.com_1758877216.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877594/media/ramp.com_1758877593.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758877294/media/ramp.com_1758877293.svg",
			},
		},
	})
}
