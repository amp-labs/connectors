package providers

const Outreach Provider = "outreach"

func init() {
	// Outreach Configuration
	SetInfo(Outreach, ProviderInfo{
		DisplayName: "Outreach",
		AuthType:    Oauth2,
		BaseURL:     "https://api.outreach.io",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.outreach.io/oauth/authorize",
			TokenURL:                  "https://api.outreach.io/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
