package providers

const Asana Provider = "asana"

func init() {
	// Asana Configuration
	SetInfo(Asana, ProviderInfo{
		DisplayName: "Asana",
		AuthType:    Oauth2,
		BaseURL:     "https://app.asana.com/api",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.asana.com/-/oauth_authorize",
			TokenURL:                  "https://app.asana.com/-/oauth_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "data.id",
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
