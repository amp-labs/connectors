package providers

const Asana Provider = "asana"

func init() {
	// Asana Configuration
	SetInfo(Asana, ProviderInfo{
		DisplayName: "Asana",
		AuthType:    Oauth2,
		BaseURL:     "https://app.asana.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163967/media/Asana_1722163967.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163991/media/Asana_1722163991.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163967/media/Asana_1722163967.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163806/media/Asana_1722163804.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.asana.com/-/oauth_authorize",
			TokenURL:                  "https://app.asana.com/-/oauth_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "data.id",
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
