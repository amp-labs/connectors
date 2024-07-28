package providers

const Airtable Provider = "airtable"

func init() {
	// Airtable Support Configuration
	SetInfo(Airtable, ProviderInfo{
		DisplayName: "Airtable",
		AuthType:    Oauth2,
		BaseURL:     "https://api.airtable.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163601/media/Airtable_1722163601.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163568/media/Airtable_1722163567.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722162786/media/Airtable_1722162786.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722163424/media/Airtable_1722163422.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 PKCE,
			AuthURL:                   "https://airtable.com/oauth2/v1/authorize",
			TokenURL:                  "https://airtable.com/oauth2/v1/token",
			ExplicitScopesRequired:    true,
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
