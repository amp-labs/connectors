package providers

const FourFour Provider = "fourFour"

func init() {
	SetInfo(FourFour, ProviderInfo{
		DisplayName: "Four/Four",
		AuthType:    Oauth2,
		BaseURL:     "https://fourfour.ai",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://fourfour.ai/oauth/authorize",
			TokenURL:                  "https://fourfour.ai/oauth/token",
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773655380/media/fourfour.ai_1773655379.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773655352/media/fourfour.ai_1773655351.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773655380/media/fourfour.ai_1773655379.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773655352/media/fourfour.ai_1773655351.svg",
			},
		},
	})
}
