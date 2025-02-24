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
				IconURL: "",
				LogoURL: "",
			},
			Regular: &MediaTypeRegular{
				IconURL: "",
				LogoURL: "",
			},
		},
	})
}
