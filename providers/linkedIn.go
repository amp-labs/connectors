package providers

const LinkedIn Provider = "linkedIn"

func init() {
	// LinkedIn configuration
	SetInfo(LinkedIn, ProviderInfo{
		DisplayName: "LinkedIn",
		AuthType:    Oauth2,
		BaseURL:     "https://api.linkedin.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:                  "https://www.linkedin.com/oauth/v2/accessToken",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
