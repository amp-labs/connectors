package providers

const PhoneBurner Provider = "phoneBurner"

func init() {
	SetInfo(PhoneBurner, ProviderInfo{
		DisplayName: "PhoneBurner",
		AuthType:    Oauth2,
		BaseURL:     "https://www.phoneburner.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			AuthURL:   "https://www.phoneburner.com/oauth/authorize",
			TokenURL:  "https://www.phoneburner.com/oauth/accesstoken",
			DocsURL:   "https://www.phoneburner.com/developer/authentication#web-flow",
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
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})
}
