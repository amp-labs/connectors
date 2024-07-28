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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722187058/media/sugarCRM_1722187057.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722187058/media/sugarCRM_1722187057.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722187058/media/sugarCRM_1722187057.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722187058/media/sugarCRM_1722187057.jpg",
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
