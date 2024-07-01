package catalog

const Gainsight Provider = "gainsight"

func init() {
	// Gainsight configuration
	SetInfo(Gainsight, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.gainsightcloud.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.gainsightcloud.com/v1/authorize",
			TokenURL:                  "https://{{.workspace}}.gainsightcloud.com/v1/users/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 AuthorizationCode,
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
