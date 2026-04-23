package providers

const Bentley Provider = "bentley"

func init() {
	// Bentley configuration
	SetInfo(Bentley, ProviderInfo{
		DisplayName: "Bentley",
		AuthType:    Oauth2,
		BaseURL:     "https://api.bentley.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://ims.bentley.com/connect/authorize",
			TokenURL:                  "https://ims.bentley.com/connect/token",
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773044816/media/bentley.com_1773044816.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773044793/media/bentley.com_1773044790.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773044816/media/bentley.com_1773044816.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773044793/media/bentley.com_1773044790.svg",
			},
		},
	})
}
