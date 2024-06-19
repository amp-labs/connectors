package providers

const Pinterest Provider = "pinterest"

func init() {
	// Pinterest Configuration
	SetInfo(Pinterest, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.pinterest.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.pinterest.com/oauth",
			TokenURL:                  "https://api.pinterest.com/v5/oauth/token",
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
