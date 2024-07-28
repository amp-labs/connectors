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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168679/media/timelyapp.com_1722168678.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168679/media/timelyapp.com_1722168678.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168679/media/timelyapp.com_1722168678.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168679/media/timelyapp.com_1722168678.jpg",
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
