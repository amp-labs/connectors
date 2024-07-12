package providers

const Typeform Provider = "typeform"

func init() {
	SetInfo(Typeform, ProviderInfo{
		DisplayName: "Typeform",
		AuthType:    Oauth2,
		BaseURL:     "https://api.typeform.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.typeform.com/oauth/authorize",
			TokenURL:                  "https://api.typeform.com/oauth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
