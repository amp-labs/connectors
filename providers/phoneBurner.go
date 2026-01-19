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
			TokenURL:  "https://www.phoneburner.com/oauth/accesstoken",
			DocsURL:   "https://www.phoneburner.com/developer/authentication#web-flow",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967555/media/paypal.com_1760967555.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967578/media/paypal.com_1760967577.svg",
			},
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})
}
