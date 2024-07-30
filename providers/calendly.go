package providers

const Calendly Provider = "calendly"

func init() {
	// Calendly Configuration
	SetInfo(Calendly, ProviderInfo{
		DisplayName: "Calendly",
		AuthType:    Oauth2,
		BaseURL:     "https://api.calendly.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.calendly.com/oauth/authorize",
			TokenURL:                  "https://auth.calendly.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346654/media/calendly_1722346653.svg", // 13
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346580/media/calendly_1722346580.svg", // 03
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346682/media/calendly_1722346682.png", // 02
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346580/media/calendly_1722346580.svg", // 11
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
