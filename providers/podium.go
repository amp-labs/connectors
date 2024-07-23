package providers

const Podium Provider = "podium"

func init() {
	SetInfo(Podium, ProviderInfo{
		DisplayName: "Podium",
		AuthType:    Oauth2,
		BaseURL:     "https://api.podium.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.podium.com/oauth/authorize",
			TokenURL:                  "https://api.podium.com/oauth/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
