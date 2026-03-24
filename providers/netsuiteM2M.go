package providers

const NetsuiteM2M Provider = "netsuiteM2M"

// Module constants are shared with the netsuite provider — same modules, different auth.

// nolint:lll,funlen
func init() {
	SetInfo(NetsuiteM2M, ProviderInfo{
		DisplayName: "Netsuite (M2M)",
		AuthType:    Custom,
		BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com",
		CustomOpts: &CustomAuthOpts{
			// No Headers or QueryParams — token acquisition is handled by the connector's
			// DynamicHeadersGenerator (JWT signing → token endpoint → Bearer header).
			Inputs: []CustomAuthInput{
				{
					Name:        "clientId",
					DisplayName: "Client ID",
					Prompt:      "From the NetSuite Integration Record (Setup > Integration > Manage Integrations).",
					DocsURL:     "https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_162686838198.html",
				},
				{
					Name:        "certificateId",
					DisplayName: "Certificate ID",
					Prompt:      "From the M2M Setup page (Setup > Integration > OAuth 2.0 Client Credentials (M2M) Setup).",
					DocsURL:     "https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_162686838198.html",
				},
				{
					Name:        "privateKey",
					DisplayName: "EC Private Key (PEM)",
					Prompt:      "The private key in PEM format (-----BEGIN EC PRIVATE KEY-----). Generate with: openssl ecparam -name prime256v1 -genkey -noout -out private.pem",
				},
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
			Read:      true,
			Subscribe: false,
			Write:     true,
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: true,
				},
			},
		},
		DefaultModule: ModuleNetsuiteRESTAPI,
		Modules: &Modules{
			ModuleNetsuiteSuiteQL: {
				DisplayName: "Netsuite M2M (SuiteQL)",
				BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/query",
				Support: Support{
					Proxy: true,
					Read:  true,
				},
			},
			ModuleNetsuiteRESTAPI: {
				DisplayName: "Netsuite M2M (REST API)",
				BaseURL:     "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/record",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
				},
			},
			ModuleNetsuiteRESTlet: {
				DisplayName: "Netsuite M2M (RESTlet)",
				BaseURL:     "https://{{.workspace}}.restlets.api.netsuite.com",
				Support: Support{
					Proxy: true,
					Read:  true,
					Write: true,
					Search: SearchSupport{
						Operators: SearchOperators{
							Equals: true,
						},
					},
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
					Prompt:      "Your NetSuite Account ID (e.g. 1234567 for production, 1234567_SB1 for sandbox).",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
				{
					Name:        "scriptId",
					DisplayName: "RESTlet Script ID",
					DocsURL:     "https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_4618456517.html",
					Prompt:      "This is an integer value for 'script' in your RESTlet's script deployment URL. If the URL is `/app/site/hosting/restlet.nl?script=3046&deploy=4`, then your script ID is `3046`.",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
				{
					Name:        "deployId",
					DisplayName: "RESTlet Deploy ID",
					DocsURL:     "https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_4618456517.html",
					Prompt:      "This is an integer value for 'deploy' in your RESTlet's script deployment URL. If the URL is `/app/site/hosting/restlet.nl?script=3046&deploy=4`, then your deploy ID is `4`.",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
			},
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "sessionTimezone",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
				{
					Name: "sessionTimezoneIsDefault",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
			},
		},
	})
}
