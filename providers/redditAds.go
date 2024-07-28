package providers

const RedditAds Provider = "redditAds"

func Init() {
	// RedditAds Configuration
	SetInfo(RedditAds, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://ads-api.reddit.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168110/media/reddit.com_1722168108.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168110/media/reddit.com_1722168108.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168110/media/reddit.com_1722168108.jpgjpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168110/media/reddit.com_1722168108.jpg",
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
