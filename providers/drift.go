package providers

const Drift Provider = "drift"

func init() {
	// Drift Configuration
	SetInfo(Drift, ProviderInfo{
		DisplayName: "Drift",
		AuthType:    Oauth2,
		BaseURL:     "https://driftapi.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dev.drift.com/authorize",
			TokenURL:                  "https://driftapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "orgId",
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
