package providers

const (
	Gorgias Provider = "gorgias"
)

func init() {
	// Gorgias Support Configuration
	SetInfo(Gorgias, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.gorgias.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.gorgias.com/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.gorgias.com/oauth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
