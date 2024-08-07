package providers

const Gong Provider = "gong"

func init() {
	// Gong configuration
	SetInfo(Gong, ProviderInfo{
		DisplayName: "Gong",
		AuthType:    Oauth2,
		BaseURL:     "https://api.gong.io",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327371/media/gong_1722327370.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327434/media/gong_1722327433.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327392/media/gong_1722327391.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327416/media/gong_1722327415.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.gong.io/oauth2/authorize",
			TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
