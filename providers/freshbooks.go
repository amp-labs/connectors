package providers

const Freshbooks Provider = "freshBooks"

func init() {
	SetInfo(Freshbooks, ProviderInfo{
		DisplayName: "Freshbooks",
		AuthType:    Oauth2,
		BaseURL:     "https://api.freshbooks.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.freshbooks.com/oauth/authorize",
			TokenURL:                  "https://api.freshbooks.com/auth/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
			DocsURL: "https://www.freshbooks.com/api/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773518149/media/freshbooks.com_1773518149.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773518227/media/freshbooks.com_1773518226.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773518195/media/freshbooks.com_1773518195.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773518106/media/freshbooks.com_1773518105.svg",
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
