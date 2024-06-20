package providers

const Aircall Provider = "aircall"

func init() {
	// Aircall Configuration
	SetInfo(Aircall, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.aircall.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dashboard.aircall.io/oauth/authorize",
			TokenURL:                  "https://api.aircall.io/v1/oauth/token",
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
