package providers

const AcuityScheduling Provider = "acuityScheduling"

func init() {
	// AcuityScheduling Configuration
	SetInfo(AcuityScheduling, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://acuityscheduling.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://acuityscheduling.com/oauth2/authorize",
			TokenURL:                  "https://acuityscheduling.com/oauth2/token",
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
