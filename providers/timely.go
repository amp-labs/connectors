package providers

const Timely Provider = "timely"

func init() {
	// Timely Configuration
	SetInfo(Timely, ProviderInfo{
		DisplayName: "Timely",
		AuthType:    Oauth2,
		BaseURL:     "https://api.timelyapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722331047/media/timely_1722331046.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722331078/media/timely_1722331078.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722331047/media/timely_1722331046.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722331097/media/timely_1722331096.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.timelyapp.com/1.1/oauth/authorize",
			TokenURL:                  "https://api.timelyapp.com/1.1/oauth/token",
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
	})
}
