package providers

const TeamleaderCRM Provider = "teamleaderCRM"

func init() {
	// TeamleaderCRM Configuration
	SetInfo(TeamleaderCRM, ProviderInfo{
		DisplayName: "Teamleader",
		AuthType:    Oauth2,
		BaseURL:     "https://api.focus.teamleader.eu",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://focus.teamleader.eu/oauth2/authorize",
			TokenURL:                  "https://focus.teamleader.eu/oauth2/access_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
