package providers

// NetsuiteM2M has the exact same functionality as Netsuite provider
// but the auth scheme is different, because Netsuite OAuth 2.0
// refresh tokens expire after 30 days, whereas M2M allows for
// longer-running integrations.
// https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_162686838198.html
const NetsuiteM2M Provider = "netsuiteM2M"

// Module constants are shared with the netsuite provider — same modules, different auth.

// nolint:lll,funlen
func init() {
	SetInfo(NetsuiteM2M, ProviderInfo{
		DisplayName: "Netsuite",
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
					DocsURL:     "https://docs.withampersand.com/customer-guides/netsuite-m2m#2-create-an-integration-record",
				},
				{
					Name:        "certificateId",
					DisplayName: "Certificate ID",
					Prompt:      "From the M2M Setup page (Setup > Integration > OAuth 2.0 Client Credentials (M2M) Setup).",
					DocsURL:     "https://docs.withampersand.com/customer-guides/netsuite-m2m#4-create-a-machine-to-machine-certificate-mapping",
				},
				{
					Name:        "privateKey",
					DisplayName: "Base64 encoded Private Key (PEM)",
					Prompt:      "The base64 encoded private key",
					DocsURL:     "https://docs.withampersand.com/customer-guides/netsuite-m2m#3-generate-a-certificate-key-pair",
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
					Prompt:      "Your NetSuite Account ID (e.g. TD1234567 for production, TD1234567_SB1 for sandbox).",
					ModuleDependencies: &ModuleDependencies{
						ModuleNetsuiteRESTAPI: ModuleDependency{},
						ModuleNetsuiteSuiteQL: ModuleDependency{},
						ModuleNetsuiteRESTlet: ModuleDependency{},
					},
				},
				{
					Name:         "scriptURL",
					DisplayName:  "RESTlet Deployment URL",
					DocsURL:      "https://docs.withampersand.com/customer-guides/netsuite-m2m#6-verify-the-deployment",
					Prompt:       "After you install the Netsuite bundle, go to the Deployments tab and copy the URL.",
					DefaultValue: "/app/site/hosting/restlet.nl?script=3277&deploy=1",
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
