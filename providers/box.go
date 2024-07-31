package providers

const Box Provider = "box"

func init() {
	// Box Configuration
	SetInfo(Box, ProviderInfo{
		DisplayName: "Box",
		AuthType:    Oauth2,
		BaseURL:     "https://api.box.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://account.box.com/api/oauth2/authorize",
			TokenURL:                  "https://api.box.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407417/media/const%20Box%20Provider%20%3D%20%22box%22_1722407417.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407337/media/const%20Box%20Provider%20%3D%20%22box%22_1722407338.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407405/media/const%20Box%20Provider%20%3D%20%22box%22_1722407406.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407291/media/const%20Box%20Provider%20%3D%20%22box%22_1722407291.svg",
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
