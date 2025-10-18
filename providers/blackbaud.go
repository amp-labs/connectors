package providers

const Blackbaud Provider = "blackbaud"

func init() {
	// Blackbaud configuration
	SetInfo(Blackbaud, ProviderInfo{
		DisplayName: "Blackbaud",
		AuthType:    Oauth2,
		BaseURL:     "https://api.sky.blackbaud.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.blackbaud.com/oauth/authorize",
			TokenURL:                  "https://oauth2.sky.blackbaud.com/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753783203/media/blackbaud.com_1753783202.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753783228/media/blackbaud.com_1753783228.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753783203/media/blackbaud.com_1753783202.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753783228/media/blackbaud.com_1753783228.png",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "bbApiSubscriptionKey",
					DisplayName: "Blackbaud API subscription key",
				},
			},
		},
	})
}
