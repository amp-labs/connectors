package providers

const Attio Provider = "attio"

func init() {
	// Attio Configuration
	SetInfo(Attio, ProviderInfo{
		DisplayName: "Attio",
		AuthType:    Oauth2,
		BaseURL:     "https://api.attio.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.attio.com/authorize",
			TokenURL:                  "https://app.attio.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724222434/media/cpdvxynal1iw2siaa8dl.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508138/media/const%20Attio%20Provider%20%3D%20%22attio%22_1722508139.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508007/media/const%20Attio%20Provider%20%3D%20%22attio%22_1722508008.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508086/media/const%20Attio%20Provider%20%3D%20%22attio%22_1722508087.svg",
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
