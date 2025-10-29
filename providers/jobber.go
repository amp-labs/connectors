package providers

const Jobber Provider = "jobber"

func init() {
	// Jobber configuration
	SetInfo(Jobber, ProviderInfo{
		DisplayName: "Jobber",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getjobber.com/api/graphql",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1748931888/media/getjobber.com_1748931887.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1748931617/media/getjobber.com_1748931616.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1748931577/Jobber_fav_nn09ua.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1748931617/media/getjobber.com_1748931616.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.getjobber.com/api/oauth/authorize",
			TokenURL:                  "https://api.getjobber.com/api/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
