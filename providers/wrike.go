package providers

const Wrike Provider = "wrike"

func init() {
	// Wrike configuration
	SetInfo(Wrike, ProviderInfo{
		DisplayName: "Wrike",
		AuthType:    Oauth2,
		BaseURL:     "https://www.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471561/media/wrike_1722471561.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471516/media/wrike_1722471514.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471586/media/wrike_1722471585.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471537/media/wrike_1722471536.svg",
			},
		},
	})
}
