package providers

const SnapchatAds Provider = "snapchatAds"

func init() {
	// Snapchat Ads configuration file
	SetInfo(SnapchatAds, ProviderInfo{
		DisplayName: "Snapchat Ads",
		AuthType:    Oauth2,
		BaseURL:     "https://adsapi.snapchat.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168560/media/snapchat.com_1722168559.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168560/media/snapchat.com_1722168559.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168560/media/snapchat.com_1722168559.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168560/media/snapchat.com_1722168559.jpg",
			},
		},
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
