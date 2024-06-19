package providers

const ZohoCRM Provider = "zohoCRM"

func init() {
	// ZohoCRM configuration
	SetInfo(ZohoCRM, ProviderInfo{
		DisplayName: "Zoho CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zohoapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_domain",
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
