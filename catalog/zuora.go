package catalog

const Zuora Provider = "zuora"

func init() {
	// Zuora Configuration
	SetInfo(Zuora, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.zuora.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://{{.workspace}}.zuora.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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
