package catalog

const Wrike Provider = "wrike"

func init() {
	// Wrike configuration
	SetInfo(Wrike, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://www.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
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
