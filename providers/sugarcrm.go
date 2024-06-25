package providers

const SugarCRM Provider = "sugarCRM"

func init() {
	// 2-legged auth
	SetInfo(SugarCRM, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 Password,
			TokenURL:                  "{{.workspace}}/rest/{{.restVersion}}/oauth2/token",
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
