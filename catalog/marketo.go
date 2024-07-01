package catalog

const Marketo Provider = "marketo"

func init() {
	// Marketo configuration file
	// workspace maps to marketo instance
	SetInfo(Marketo, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.mktorest.com",
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
