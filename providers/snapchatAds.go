package providers

const SnapchatAds Provider = "snapchatAds"

func init() {
	// Snapchat Ads configuration file
	SetInfo(SnapchatAds, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://adsapi.snapchat.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://accounts.snapchat.com/login/oauth2/authorize",
			TokenURL:                  "https://accounts.snapchat.com/login/oauth2/access_token",
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
