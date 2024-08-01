package providers

const ClickUp Provider = "clickup"

func init() {
	// ClickUp Support Configuration
	SetInfo(ClickUp, ProviderInfo{
		DisplayName: "ClickUp",
		AuthType:    Oauth2,
		BaseURL:     "https://api.clickup.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.clickup.com/api",
			TokenURL:                  "https://api.clickup.com/api/v2/oauth/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722537393/media/clickup.com_1722537393.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722537424/media/clickup.com_1722537424.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722536894/media/clickup.com_1722536893.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722537034/media/clickup.com_1722537033.svg",
			},
		},
	})
}
