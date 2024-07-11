package providers

const CampaignMonitor Provider = "campaignMonitor"

func init() {
	// CampaignMonitor Configuration
	SetInfo(CampaignMonitor, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.createsend.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.createsend.com/oauth",
			TokenURL:                  "https://api.createsend.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
