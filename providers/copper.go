package providers

const Copper Provider = "copper"

func init() {
	// Copper configuration
	SetInfo(Copper, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.copper.com/developer_api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.copper.com/oauth/authorize",
			TokenURL:                  "https://app.copper.com/oauth/token",
			ExplicitScopesRequired:    true,
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
