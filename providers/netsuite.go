package providers

const (
	Netsuite Provider = "netsuite"

	// ModuleNetsuiteSuiteQL is a read-only module that uses SuiteQL to read data.
	ModuleNetsuiteSuiteQL = "suiteql"

	// ModuleNetsuiteRESTAPI is a read-write module that uses the REST API to read and write data.
	ModuleNetsuiteRESTAPI = "restapi"
)

// nolint:lll,funlen
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		DefaultModule: ModuleNetsuiteRESTAPI,
		Modules: &Modules{
			ModuleNetsuiteSuiteQL: {
				DisplayName: "Netsuite (SuiteQL)",
				BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/query",
				Support: Support{
					Proxy: true,
					Read:  true,
				},
			},
			ModuleNetsuiteRESTAPI: {
				DisplayName: "Netsuite (REST API)",
				BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/record",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Account ID",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
					},
				},
			},
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "sessionTimezone",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
					},
				},
			},
		},
	})
}
