package providers

const GetResponse Provider = "getResponse"

func init() {
	// GetResponse configuration

	SetInfo(GetResponse, ProviderInfo{
		DisplayName: "GetResponse",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getresponse.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.getresponse.com/oauth2_authorize.html",
			TokenURL:                  "https://api.getresponse.com/v3/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326298/media/const%20GetResponse%20Provider%20%3D%20%22getResponse%22_1722326298.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326391/media/const%20GetResponse%20Provider%20%3D%20%22getResponse%22_1722326391.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326298/media/const%20GetResponse%20Provider%20%3D%20%22getResponse%22_1722326298.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326361/media/const%20GetResponse%20Provider%20%3D%20%22getResponse%22_1722326361.svg",
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
