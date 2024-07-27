package providers

const StackExchange Provider = "stackExchange"

func init() {
	// StackExchange configuration
	SetInfo(StackExchange, ProviderInfo{
		DisplayName: "StackExchange",
		AuthType:    Oauth2,
		BaseURL:     "https://api.stackexchange.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://stackoverflow.com/oauth",
			TokenURL:                  "https://stackoverflow.com/oauth/access_token/json",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062606/media/stackExchange_1722062605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062568/media/stackExchange_1722062567.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062606/media/stackExchange_1722062605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062537/media/stackExchange_1722062535.svg",
			},
		},
	})
}
