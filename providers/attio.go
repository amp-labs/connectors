package providers

const Attio Provider = "attio"

func init() {
	// Attio Configuration
	SetInfo(Attio, ProviderInfo{
		DisplayName: "Attio",
		AuthType:    Oauth2,
		BaseURL:     "https://api.attio.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.attio.com/authorize",
			TokenURL:                  "https://app.attio.com/oauth/token",
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
