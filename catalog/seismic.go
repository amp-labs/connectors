package catalog

const Seismic Provider = "seismic"

func init() {
	SetInfo(Seismic, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.seismic.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.seismic.com/tenants/{{.workspace}}/connect/authorize",
			TokenURL:                  "https://auth.seismic.com/tenants/{{.workspace}}/connect/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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
