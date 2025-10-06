package providers

const Copper Provider = "copper"

func init() {
	// Copper configuration
	SetInfo(Copper, ProviderInfo{
		DisplayName: "Copper",
		AuthType:    Oauth2,
		BaseURL:     "https://api.copper.com/developer_api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.copper.com/oauth/authorize",
			TokenURL:                  "https://app.copper.com/oauth/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/f7mytk1fsugjgukq6s2i.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/f7mytk1fsugjgukq6s2i.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478129/media/copper_1722478128.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478080/media/copper_1722478079.svg",
			},
		},
	})
}
