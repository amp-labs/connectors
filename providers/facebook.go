package providers

const Facebook Provider = "facebook"

func init() {
	// Facebook Ads Manager Configuration
	SetInfo(Facebook, ProviderInfo{
		DisplayName: "Facebook",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.facebook.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.facebook.com/v19.0/dialog/oauth",
			TokenURL:                  "https://graph.facebook.com/v19.0/oauth/access_token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478709/media/facebook_1722478708.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478689/media/facebook_1722478688.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478709/media/facebook_1722478708.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722478689/media/facebook_1722478688.svg",
			},
		},
	})
}
