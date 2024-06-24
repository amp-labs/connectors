package catalog

const StackExchange Provider = "stackExchange"

func init() {
	// StackExchange configuration
	SetInfo(StackExchange, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.stackexchange.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://stackoverflow.com/oauth",
			TokenURL:                  "https://stackoverflow.com/oauth/access_token/json",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
