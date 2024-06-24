package catalog

const Intercom Provider = "intercom"

func init() {
	// Intercom configuration
	SetInfo(Intercom, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.intercom.io",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.intercom.com/oauth",
			TokenURL:                  "https://api.intercom.io/auth/eagle/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
