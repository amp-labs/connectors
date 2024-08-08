package providers

const Basecamp Provider = "basecamp"

func init() {
	// Basecamp Configuration
	// The wokspace in baseURL should be mapped to accounts.id
	SetInfo(Basecamp, ProviderInfo{
		DisplayName: "Basecamp",
		AuthType:    Oauth2,
		BaseURL:     "https://3.basecampapi.com/{{.workspace}}",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324615/media/basecamp_1722324614.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324674/media/basecamp_1722324673.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324615/media/basecamp_1722324614.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324674/media/basecamp_1722324673.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://launchpad.37signals.com/authorization/new?type=web_server",
			TokenURL:                  "https://launchpad.37signals.com/authorization/token?type=refresh",
			ExplicitScopesRequired:    false,
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
	})
}
