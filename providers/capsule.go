package providers

const Capsule Provider = "capsule"

func init() {
	// Capsule Configuration
	SetInfo(Capsule, ProviderInfo{
		DisplayName: "Capsule",
		AuthType:    Oauth2,
		BaseURL:     "https://api.capsulecrm.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.capsulecrm.com/oauth/authorise",
			TokenURL:                  "https://api.capsulecrm.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722509743/media/const%20Capsule%20Provider%20%3D%20%22capsule%22_1722509744.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722509768/media/const%20Capsule%20Provider%20%3D%20%22capsule%22_1722509769.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722509743/media/const%20Capsule%20Provider%20%3D%20%22capsule%22_1722509744.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722509768/media/const%20Capsule%20Provider%20%3D%20%22capsule%22_1722509769.svg",
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
