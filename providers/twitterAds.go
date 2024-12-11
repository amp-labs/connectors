// nolint:lll
package providers

const TwitterAds Provider = "twitterAds"

func init() { // nolint:funlen
	SetInfo(TwitterAds, ProviderInfo{
		DisplayName: "Twitter Ads",
		AuthType:    Oauth2,
		BaseURL:     "https://ads-api.x.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://twitter.com/i/oauth2/authorize",
			TokenURL:                  "https://api.twitter.com/2/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733936359/media/twitterAds_1733936358.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733936359/media/twitterAds_1733936358.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733936180/media/twitterAds_1733936179.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733936180/media/twitterAds_1733936179.svg",
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
