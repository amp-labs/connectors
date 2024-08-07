package providers

const Mural Provider = "mural"

func init() {
	// Mural Configuration
	SetInfo(Mural, ProviderInfo{
		DisplayName: "Mural",
		AuthType:    Oauth2,
		BaseURL:     "https://api.mural.co/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.mural.co/oauth/authorize",
			TokenURL:                  "https://api.mural.co/oauth/token",
			ExplicitScopesRequired:    true,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469525/media/mural_1722469525.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469499/media/mural_1722469498.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469525/media/mural_1722469525.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722469499/media/mural_1722469498.svg",
			},
		},
	})
}
