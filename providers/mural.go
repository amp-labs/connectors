package providers

const Mural Provider = "mural"

func init() {
	// Mural Configuration
	SetInfo(Mural, ProviderInfo{
		DisplayName: "Mural",
		AuthType:    Oauth2,
		BaseURL:     "https://app.mural.co/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.mural.co/api/public/v1/authorization/oauth2/",
			TokenURL:                  "https://app.mural.co/api/public/v1/authorization/oauth2/token",
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
