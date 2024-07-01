package catalog

const Klaviyo Provider = "klaviyo"

func init() {
	// Klaviyo configuration
	SetInfo(Klaviyo, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://a.klaviyo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 PKCE,
			AuthURL:                   "https://www.klaviyo.com/oauth/authorize",
			TokenURL:                  "https://a.klaviyo.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
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
