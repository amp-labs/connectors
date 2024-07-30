package providers

const Teamwork Provider = "teamwork"

func init() {
	// Teamwork Configuration
	SetInfo(Teamwork, ProviderInfo{
		DisplayName: "Teamwork",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.teamwork.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.teamwork.com/launchpad/login",
			TokenURL:                  "https://www.teamwork.com/launchpad/v1/token.json",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user.id",
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
