package catalog

const Atlassian Provider = "atlassian"

func init() {
	// Atlassian Configuration
	SetInfo(Atlassian, ProviderInfo{
		DisplayName: "Atlassian Jira",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
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
