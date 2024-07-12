package providers

const AWeber Provider = "aWeber"

func init() {
	// AWeber Configuration
	SetInfo(AWeber, ProviderInfo{
		DisplayName: "AWeber",
		AuthType:    Oauth2,
		BaseURL:     "https://api.aweber.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://auth.aweber.com/oauth2/authorize",
			TokenURL:                  "https://auth.aweber.com/oauth2/token",
			ExplicitScopesRequired:    true,
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
