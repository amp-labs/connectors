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
			AuthURLParams: map[string]string{
				// If acting on behalf of a vendor account, PhoneBurner requires owner_type=vendor.
				// Docs: https://www.phoneburner.com/developer/authentication#web-flow
				"owner_type": "vendor",
			},
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768827171/media/phoneburner.com_1768827171.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1768827171/media/phoneburner.com_1768827171.svg",
			},
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})
}
