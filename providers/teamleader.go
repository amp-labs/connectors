package providers

const Teamleader Provider = "teamleader"

func init() {
	// Teamleader Configuration
	SetInfo(Teamleader, ProviderInfo{
		DisplayName: "Teamleader",
		AuthType:    Oauth2,
		BaseURL:     "https://api.focus.teamleader.eu",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://focus.teamleader.eu/oauth2/authorize",
			TokenURL:                  "https://focus.teamleader.eu/oauth2/access_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722418524/media/const%20Teamleader%20Provider%20%3D%20%22teamleader%22_1722418523.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722418524/media/const%20Teamleader%20Provider%20%3D%20%22teamleader%22_1722418523.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722418524/media/const%20Teamleader%20Provider%20%3D%20%22teamleader%22_1722418523.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722418524/media/const%20Teamleader%20Provider%20%3D%20%22teamleader%22_1722418523.png",
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
