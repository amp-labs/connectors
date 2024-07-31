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
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722433816/media/const%20ClickUp%20Provider%20%3D%20%22clickup%22_1722433817.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722433753/media/const%20ClickUp%20Provider%20%3D%20%22clickup%22_1722433753.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722433816/media/const%20ClickUp%20Provider%20%3D%20%22clickup%22_1722433817.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722433776/media/const%20ClickUp%20Provider%20%3D%20%22clickup%22_1722433777.svg",
			},
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
	})
}
