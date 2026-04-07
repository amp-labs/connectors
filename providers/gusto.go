package providers

const Gusto Provider = "gusto"

func init() {
	SetInfo(Gusto, ProviderInfo{
		DisplayName: "Gusto",
		AuthType:    Oauth2,
		BaseURL:     "https://api.gusto.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.gusto.com/oauth/authorize",
			TokenURL:                  "https://api.gusto.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1775077438/media/gusto.com_1775077438.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1775077356/media/gusto.com_1775077354.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1775077438/media/gusto.com_1775077438.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1775077412/media/gusto.com_1775077412.svg",
			},
		},
	})
}
