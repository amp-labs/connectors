package providers

import (
	"net/http"
)

// SalesforceCustomDomain is a twin of the Salesforce provider for orgs that are
// not reached at their my.salesforce.com domain. It targets the same APIs and
// modules, so the connector implementation in providers/salesforce is reused
// under a different provider name (see WithProvider in that package).
//
// Motivation: the Salesforce entry hardcodes the API and OAuth hosts as
// {{.workspace}}.my.salesforce.com. Some orgs cannot be reached that way — an
// enterprise egress policy may require API traffic to traverse a gateway, and
// some builders need OAuth to run against login.salesforce.com so that SSO
// users can authenticate. This entry parameterizes both hosts independently so
// each can be pointed wherever a given connection requires.
const SalesforceCustomDomain Provider = "salesforceCustomDomain"

// nolint:funlen
func init() {
	SetInfo(SalesforceCustomDomain, ProviderInfo{
		DisplayName: "Salesforce (Custom Domain)",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.apiDomain}}",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.authDomain}}/services/oauth2/userinfo",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCodePKCE,
			AuthURL:                "https://{{.authDomain}}/services/oauth2/authorize",
			AuthURLParams:          map[string]string{"prompt": "login"},
			TokenURL:               "https://{{.authDomain}}/services/oauth2/token",
			ExplicitScopesRequired: false,
			// The API host is supplied as metadata rather than derived from the
			// org, so it must be collected before the OAuth flow begins.
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "id",
				ScopesField:      "scope",
				// WorkspaceRefField is deliberately unset. Salesforce returns
				// instance_url pointing at the org's own domain, which is not
				// where this connection sends requests, so adopting it as the
				// workspace ref would contradict apiDomain.
			},
		},
		DefaultModule: ModuleSalesforceCRM,
		Modules: &Modules{
			ModuleSalesforceCRM: {
				BaseURL:     "https://{{.apiDomain}}",
				DisplayName: "Salesforce (Custom Domain)",
				Support: Support{
					BatchWrite: &BatchWriteSupport{
						Create: BatchWriteSupportConfig{
							DefaultRecordLimit: new(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
						Update: BatchWriteSupportConfig{
							DefaultRecordLimit: new(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
					},
					BulkWrite: BulkWriteSupport{
						Insert: false,
						Update: false,
						Upsert: true,
						Delete: true,
					},
					Delete:    true,
					Proxy:     true,
					Read:      true,
					Subscribe: true,
					Write:     true,
					Search: SearchSupport{
						Operators: SearchOperators{
							Equals: true,
						},
					},
				},
				SubscribeRequirements: &SubscribeRequirements{
					Registration:   new(true),
					PostProcess:    new(true),
					SubscribeByAPI: new(true),
				},
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: true,
				Delete: true,
			},
			Delete:    true,
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: true,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "apiDomain",
					DisplayName: "API domain",
					Prompt: "Host that serves the Salesforce REST, Bulk and Tooling APIs for this org, " +
						"including any path prefix (e.g. `acme.my.salesforce.com`, or " +
						"`gateway.example.com/salesforce` when traffic is routed through a gateway).",
					ModuleDependencies: &ModuleDependencies{
						ModuleSalesforceCRM: {},
					},
				},
				{
					Name:        "authDomain",
					DisplayName: "OAuth domain",
					// Defaulting here also means an unset value never reaches the
					// catalog templates, which are resolved with missingkey=error.
					DefaultValue: "login.salesforce.com",
					Prompt: "Host that serves the OAuth authorize and token endpoints. " +
						"Defaults to login.salesforce.com, which works for most orgs including SSO users.",
					DocsURL: "https://help.salesforce.com/s/articleView?id=000388956&type=1",
					ModuleDependencies: &ModuleDependencies{
						ModuleSalesforceCRM: {},
					},
				},
			},
		},
	})
}
