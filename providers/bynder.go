package providers

const Bynder Provider = "bynder"

func init() {
	SetInfo(Bynder, ProviderInfo{
		DisplayName: "bynder",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.bynder.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.bynder.com/v6/authentication/oauth2/auth",
			TokenURL:                  "https://{{.workspace}}.bynder.com/v6/authentication/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
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
