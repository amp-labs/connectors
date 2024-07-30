package providers

const Gainsight Provider = "gainsight"

func init() {
	// Gainsight configuration
	SetInfo(Gainsight, ProviderInfo{
		DisplayName: "Gainsight",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.gainsightcloud.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.gainsightcloud.com/v1/authorize",
			TokenURL:                  "https://{{.workspace}}.gainsightcloud.com/v1/users/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 AuthorizationCode,
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326070/media/const%20Gainsight%20Provider%20%3D%20%22gainsight%22_1722326070.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326115/media/const%20Gainsight%20Provider%20%3D%20%22gainsight%22_1722326114.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326012/media/const%20Gainsight%20Provider%20%3D%20%22gainsight%22_1722326012.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326150/media/const%20Gainsight%20Provider%20%3D%20%22gainsight%22_1722326150.svg",
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
