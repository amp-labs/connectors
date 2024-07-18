package providers

const Webflow Provider = "webflow"

func init() {
	// Webflow Support Configuration
	SetInfo(Webflow, ProviderInfo{
		DisplayName: "Webflow",
		AuthType:    Oauth2,
		BaseURL:     "https://api.webflow.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://webflow.com/oauth/authorize",
			TokenURL:                  "https://api.webflow.com/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
