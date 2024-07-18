package providers

const Instagram Provider = "instagram"

func init() {
	// Instagram Configuration
	// TODO: Supports only short-lived tokens
	SetInfo(Instagram, ProviderInfo{
		DisplayName: "Instagram",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.instagram.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.instagram.com/oauth/authorize",
			TokenURL:                  "https://api.instagram.com/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "user_id",
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
