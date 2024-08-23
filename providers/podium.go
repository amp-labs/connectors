package providers

const Podium Provider = "podium"

func init() {
	SetInfo(Podium, ProviderInfo{
		DisplayName: "Podium",
		AuthType:    Oauth2,
		BaseURL:     "https://api.podium.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.podium.com/oauth/authorize",
			TokenURL:                  "https://api.podium.com/oauth/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724222553/media/nxtssrgengo6pbbwqwd2.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330479/media/podium_1722330478.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724365495/media/cbirwotlb7si9qrdicok.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330504/media/podium_1722330503.svg",
			},
		},
	})
}
