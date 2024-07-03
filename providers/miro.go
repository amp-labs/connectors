package providers

const Miro Provider = "miro"

func init() {
	// Miro Support Configuration
	SetInfo(Miro, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.miro.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://miro.com/oauth/authorize",
			TokenURL:                  "https://api.miro.com/v1/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "user_id",
				WorkspaceRefField: "team_id",
				ScopesField:       "scope",
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
