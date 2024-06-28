package providers

const RedditAds Provider = "redditAds"

func Init() {
	// RedditAds Configuration
	SetInfo(RedditAds, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://ads-api.reddit.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.reddit.com/api/v1/authorize",
			TokenURL:                  "https://www.reddit.com/api/v1/access_token",
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
