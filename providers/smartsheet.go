package providers

const Smartsheet Provider = "smartsheet"

func init() {
	// Smartsheet Support Configuration
	SetInfo(Smartsheet, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.smartsheet.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.smartsheet.com/b/authorize",
			TokenURL:                  "https://api.smartsheet.com/2.0/token",
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
