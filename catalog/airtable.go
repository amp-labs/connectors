package catalog

const Airtable Provider = "airtable"

func init() {
	// Airtable Support Configuration
	SetInfo(Airtable, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.airtable.com",
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
