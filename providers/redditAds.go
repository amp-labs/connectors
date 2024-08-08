package providers

const RedditAds Provider = "redditAds"

func Init() {
	// RedditAds Configuration
	SetInfo(RedditAds, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://ads-api.reddit.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329788/media/reddit_1722329787.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329864/media/reddit_1722329863.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329788/media/reddit_1722329787.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329840/media/reddit_1722329840.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.reddit.com/api/v1/authorize",
			AuthURLParams:             map[string]string{"duration": "permanent"},
			TokenURL:                  "https://www.reddit.com/api/v1/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DisplayName: "Reddit Ads",
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
