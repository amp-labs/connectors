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
	})
}
