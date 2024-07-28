package providers

const Aha Provider = "aha"

func init() {
	// Aha Configuration
	SetInfo(Aha, ProviderInfo{
		DisplayName: "Aha",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.aha.io/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.aha.io/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.aha.io/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184261/media/aha_1722184260.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184261/media/aha_1722184260.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184261/media/aha_1722184260.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184261/media/aha_1722184260.jpg",
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
