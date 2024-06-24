package catalog

const WordPress Provider = "wordPress"

func init() {
	// WordPress Support configuration
	SetInfo(WordPress, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://public-api.wordpress.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
