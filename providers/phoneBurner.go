package providers

const PhoneBurner Provider = "phoneBurner"

func init() {
	SetInfo(PhoneBurner, ProviderInfo{
		DisplayName: "PhoneBurner",
		AuthType:    Oauth2,
		BaseURL:     "https://www.phoneburner.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			AuthURL:   "https://www.phoneburner.com/oauth/authorize",
			TokenURL:                  "https://www.phoneburner.com/oauth/accesstoken",
			DocsURL:                   "https://www.phoneburner.com/developer/authentication#web-flow",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1769032944/phoneBurner_icon_j5zikt.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1769033375/phoneBurner_logo_dark_mxc9lt.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1769032944/phoneBurner_icon_j5zikt.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768827171/media/phoneburner.com_1768827171.svg",
			},
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})
}
