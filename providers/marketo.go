package providers

const Marketo Provider = "marketo"

func init() {
	// Marketo configuration file
	// workspace maps to marketo instance
	SetInfo(Marketo, ProviderInfo{
		DisplayName: "Marketo",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.mktorest.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166604/media/marketo.com_1722166602.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166604/media/marketo.com_1722166602.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166604/media/marketo.com_1722166602.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166604/media/marketo.com_1722166602.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			TokenURL:                  "https://{{.workspace}}.mktorest.com/identity/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 ClientCredentials,
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
