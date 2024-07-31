package providers

const Domo Provider = "domo"

func init() {
	// Domo configuration file
	SetInfo(Domo, ProviderInfo{
		DisplayName: "Domo",
		AuthType:    Oauth2,
		BaseURL:     "https://api.domo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://api.domo.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "userId",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722455548/media/domo_1722455546.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722455548/media/domo_1722455546.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722455369/media/domo_1722455368.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722455369/media/domo_1722455368.svg",
			},
		},
	})
}
