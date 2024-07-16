package providers

const Hubspot Provider = "hubspot"

func init() {
	// Hubspot configuration
	SetInfo(Hubspot, ProviderInfo{
		DisplayName: "HubSpot",
		AuthType:    Oauth2,
		BaseURL:     "https://api.hubapi.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.hubspot.com/oauth/authorize",
			TokenURL:                  "https://api.hubapi.com/oauth/v1/token",
			ExplicitScopesRequired:    true,
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
