package providers

const Drip Provider = "drip"

func init() {
	// Drip configuration
	SetInfo(Drip, ProviderInfo{
		DisplayName: "Drip",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getdrip.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.getdrip.com/oauth/authorize",
			TokenURL:                  "https://www.getdrip.com/oauth/token",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731568403/media/drip.com_1731568403.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731568457/media/drip.com_1731568458.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731568403/media/drip.com_1731568403.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731568430/media/drip.com_1731568431.svg",
			},
		},
	})
}
