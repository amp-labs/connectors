package catalog

const Capsule Provider = "capsule"

func init() {
	// Capsule Configuration
	SetInfo(Capsule, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.capsulecrm.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.capsulecrm.com/oauth/authorise",
			TokenURL:                  "https://api.capsulecrm.com/oauth/token",
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
