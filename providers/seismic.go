package providers

const Seismic Provider = "seismic"

func init() {
	SetInfo(Seismic, ProviderInfo{
		DisplayName: "Seismic",
		AuthType:    Oauth2,
		BaseURL:     "https://api.seismic.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.seismic.com/tenants/{{.workspace}}/connect/authorize",
			TokenURL:                  "https://auth.seismic.com/tenants/{{.workspace}}/connect/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348404/media/seismic_1722348404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348429/media/seismic_1722348428.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348404/media/seismic_1722348404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348448/media/seismic_1722348447.svg",
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
