package providers

const ServiceNow Provider = "serviceNow"

func init() {
	// ServiceNow configuration
	SetInfo(ServiceNow, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.service-now.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.service-now.com/oauth_auth.do",
			TokenURL:                  "https://{{.workspace}}.service-now.com/oauth_token.do",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
