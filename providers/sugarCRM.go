package providers

const SugarCRM Provider = "sugarCRM"

func init() {
	// 2-legged auth
	SetInfo(SugarCRM, ProviderInfo{
		DisplayName: "SugarCRM",
		AuthType:    Oauth2,
		BaseURL:     "{{.workspace}}/rest/v11_24",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 Password,
			TokenURL:                  "{{.workspace}}/rest/v11_24/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348686/media/sugarCRM_1722348686.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348716/media/sugarCRM_1722348716.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348686/media/sugarCRM_1722348686.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348716/media/sugarCRM_1722348716.svg",
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
