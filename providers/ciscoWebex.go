package providers

const CiscoWebex Provider = "ciscoWebex"

func init() {
	SetInfo(CiscoWebex, ProviderInfo{
		DisplayName: "Cisco Webex",
		AuthType:    Oauth2,
		BaseURL:     "https://webexapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://webexapis.com/v1/authorize",
			TokenURL:                  "https://webexapis.com/v1/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
			DocsURL: "https://developer.webex.com/docs/run-an-oauth-integration",
		},

		// Media: Media assets (logos/icons) will be added in a follow-up PR
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
