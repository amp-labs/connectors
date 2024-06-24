package catalog

const ClickUp Provider = "clickup"

func init() {
	// ClickUp Support Configuration
	SetInfo(ClickUp, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.clickup.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.clickup.com/api",
			TokenURL:                  "https://api.clickup.com/api/v2/oauth/token",
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
