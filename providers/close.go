package providers

const Close Provider = "close"

func init() {
	// Close Configuration
	SetInfo(Close, ProviderInfo{
		DisplayName: "Close",
		AuthType:    Oauth2,
		BaseURL:     "https://api.close.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.close.com/oauth2/authorize",
			TokenURL:                  "https://api.close.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "user_id",
				WorkspaceRefField: "organization_id",
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
