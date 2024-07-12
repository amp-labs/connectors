package providers

const Keap Provider = "keap"

func init() {
	// Keap configuration
	SetInfo(Keap, ProviderInfo{
		DisplayName: "Keap",
		AuthType:    Oauth2,
		BaseURL:     "https://api.infusionsoft.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.infusionsoft.com/app/oauth/authorize",
			TokenURL:                  "https://api.infusionsoft.com/token",
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
