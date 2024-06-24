package catalog

const Slack Provider = "slack"

func init() {
	// Slack configuration
	SetInfo(Slack, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://slack.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://slack.com/oauth/v2/authorize",
			TokenURL:                  "https://slack.com/api/oauth.v2.access",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:       "scope",
				WorkspaceRefField: "workspace_name",
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
