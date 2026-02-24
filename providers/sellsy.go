package providers

const Sellsy Provider = "sellsy"

func init() {
	// Sellsy configuration
	SetInfo(Sellsy, ProviderInfo{
		DisplayName: "Sellsy",
		AuthType:    Oauth2,
		BaseURL:     "https://api.sellsy.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://login.sellsy.com/oauth2/authorization",
			TokenURL:                  "https://login.sellsy.com/oauth2/access-tokens",
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
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470945/media/sellsy_1722470945.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471161/media/sellsy_1722471161.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470988/media/sellsy_1722470988.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471227/media/sellsy_1722471226.svg",
			},
		},
	})
}
