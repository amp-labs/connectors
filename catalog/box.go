package catalog

const Box Provider = "box"

func init() {
	// Box Configuration
	SetInfo(Box, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.box.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://account.box.com/api/oauth2/authorize",
			TokenURL:                  "https://api.box.com/oauth2/token",
			ExplicitScopesRequired:    false,
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
