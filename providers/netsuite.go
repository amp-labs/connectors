package providers

const Netsuite Provider = "netsuite"

func init() {
	SetInfo(Netsuite, ProviderInfo{
		DisplayName: "Netsuite",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.app.netsuite.com/app/login/oauth2/authorize.nl",
			TokenURL:                  "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/auth/oauth2/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			DocsURL:                   "https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_157771733782.html#procedure_157838925981",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740403711/media/netsuite_1740403705.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740411027/netsuite_xtpygf.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740403711/media/netsuite_1740403705.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740404009/media/netsuite_1740403997.svg",
			},
		},
	})
}
