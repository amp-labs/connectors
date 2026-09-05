package providers

import (
	"net/http"
)

// SalesforceCustomDomain is a twin of the Salesforce provider for orgs whose
// APIs are not reached at their my.salesforce.com domain. It targets the same
// APIs and modules, so the connector implementation in providers/salesforce is
// reused under a different provider name (see WithProvider in that package).
//
// Motivation: the Salesforce entry hardcodes every host as
// {{.workspace}}.my.salesforce.com. An enterprise egress policy may instead
// require all Salesforce traffic — data and token alike — to traverse a
// customer-operated gateway, reachable at an arbitrary host and path prefix.
//
// The workspace therefore carries that whole prefix rather than a subdomain,
// and both the token endpoint and the API base derive from it. Keeping it in
// the workspace rather than a bespoke variable matters: the substitution map
// always carries a workspace, so resolving ProviderInfo without a connection —
// as the provider-metadata endpoint does — degrades to an empty host instead of
// failing on a missing key.
//
// The grant is client credentials because such a gateway typically fronts the
// token endpoint but not the interactive authorize page, and because these
// connections are server-to-server with no consumer present to complete a
// browser redirect.
const SalesforceCustomDomain Provider = "salesforceCustomDomain"

// nolint:funlen
func init() {
	SetInfo(SalesforceCustomDomain, ProviderInfo{
		DisplayName: "Salesforce (Custom Domain)",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.workspace}}/services/oauth2/userinfo",
		},
		Oauth2Opts: &Oauth2Opts{
			// Client credentials rather than an authorization code grant: the
			// gateway fronting Salesforce serves the token endpoint but not the
			// interactive authorize page, and these connections are
			// server-to-server with no consumer present to complete a redirect.
			GrantType:              ClientCredentials,
			TokenURL:               "https://{{.workspace}}/services/oauth2/token",
			ExplicitScopesRequired: false,
			// The workspace carries the API host rather than a subdomain, so it
			// must be collected before the OAuth flow begins.
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "id",
				ScopesField:      "scope",
				// WorkspaceRefField is deliberately unset. Salesforce returns
				// instance_url pointing at the org's own domain, which is not
				// necessarily where this connection sends requests, so adopting
				// it would overwrite the API host the builder supplied.
			},
		},
		DefaultModule: ModuleSalesforceCRM,
		Modules: &Modules{
			ModuleSalesforceCRM: {
				BaseURL:     "https://{{.workspace}}",
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
					Name:        "workspace",
					DisplayName: "API domain",
					Prompt: "Host that serves the Salesforce REST, Bulk and Tooling APIs for this org, " +
						"including any path prefix (e.g. `acme.my.salesforce.com`, or " +
						"`gateway.example.com/salesforce` when traffic is routed through a gateway).",
					ModuleDependencies: &ModuleDependencies{
						ModuleSalesforceCRM: {},
					},
				},
			},
		},
	})
}
