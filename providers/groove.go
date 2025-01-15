package providers

const Groove Provider = "groove"

func init() {
	// Groove configuration
	SetInfo(Groove, ProviderInfo{
		DisplayName: "Groove",
		AuthType:    Oauth2,
		BaseURL:     "https://api.groovehq.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1736955540/media/inahuieqaf3j6jipw5l1.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1736955651/media/paozgcxtoudfmmrhokfz.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1736955682/media/rwtps5ybpjxvm6ywrz0p.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1736955709/media/iqpdotcxrdjt4ghla6at.png",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.groovehq.com/oauth/authorize",
			TokenURL:                  "https://api.groovehq.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
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
	})
}
