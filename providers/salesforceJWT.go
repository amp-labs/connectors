package providers

import (
	"net/http"

	"github.com/amp-labs/connectors/internal/goutils"
)

// SalesforceJWT is a twin of the Salesforce provider that authenticates using
// the OAuth 2.0 JWT Bearer flow (RFC 7523 §2.1) instead of the Authorization
// Code grant. It targets the same underlying Salesforce APIs and modules, so
// the connector implementation in providers/salesforce is reused under a
// different provider name (see WithProvider in that package).
//
// Motivation: the Authorization Code flow is interactive and its refresh
// tokens can be revoked by Salesforce admin policy, whereas the JWT Bearer
// flow is fully server-to-server and suitable for unattended integrations.
const SalesforceJWT Provider = "salesforceJWT"

// nolint:lll,funlen
func init() {
	SetInfo(SalesforceJWT, ProviderInfo{
		DisplayName: "Salesforce",
		AuthType:    Custom,
		BaseURL:     "https://{{.workspace}}.my.salesforce.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.workspace}}.my.salesforce.com/services/oauth2/userinfo",
		},
		CustomOpts: &CustomAuthOpts{
			// Token acquisition is handled by the connector's DynamicHeadersGenerator
			// (JWT signing → token endpoint → Bearer header). See providers/salesforce/jwt.
			Inputs: []CustomAuthInput{
				{
					Name:        "clientId",
					DisplayName: "Consumer Key",
					Prompt:      "The Consumer Key (Client ID) of the Salesforce External Client App configured for JWT Bearer flow. Found under Setup → External Client App Manager → [Your App] → Settings → OAuth Settings -> Consumer Key and Secret.",
					DocsURL:     "https://docs.withampersand.com/customer-guides/salesforce",
				},
				{
					Name:        "username",
					DisplayName: "Salesforce Username",
					Prompt:      "The Salesforce username the integration should act as (e.g. integration.user@acme.com). This becomes the JWT 'sub' claim. The user must be pre-authorized on the External Client App.",
					DocsURL:     "https://docs.withampersand.com/customer-guides/salesforce",
				},
				{
					Name:        "privateKey",
					DisplayName: "Base64 encoded Private Key (PEM)",
					Prompt:      "The base64-encoded RSA private key (PEM) whose X.509 certificate is registered on the External Client App under OAuth Settings → 'Use digital signatures'. Must be RSA — EC keys are not supported by Salesforce for this flow.",
					DocsURL:     "https://docs.withampersand.com/customer-guides/salesforce",
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
			Delete: true,
			Proxy:  true,
			Read:   true,
			// Subscribe (CDC / Event Relay) is supported in principle over JWT Bearer
			// auth — the Metadata API calls that register EventChannel/NamedCredential/
			// EventRelayConfig work over any authenticated HTTP client, and the
			// downstream auth is delegated to AWS via the NamedCredential endpoint.
			// Keep parity with the OAuth2 Salesforce provider; flip to false here if
			// live testing surfaces a JWT-specific incompatibility.
			Subscribe: true,
			Write:     true,
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: true,
				},
			},
		},
		DefaultModule: ModuleSalesforceCRM,
		Modules: &Modules{
			ModuleSalesforceCRM: {
				BaseURL:     "https://{{.workspace}}.my.salesforce.com",
				DisplayName: "Salesforce",
				Support: Support{
					BatchWrite: &BatchWriteSupport{
						Create: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
						Update: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
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
					DisplayName: "Subdomain",
					DocsURL:     "https://help.salesforce.com/s/articleView?language=en_US&id=sf.faq_domain_name_what.htm&type=5",
					Prompt:      "Your Salesforce My Domain subdomain (e.g. `acme` for acme.my.salesforce.com, or `acme--dev.sandbox` for sandbox).",
					ModuleDependencies: &ModuleDependencies{
						ModuleSalesforceCRM: {},
					},
				},
			},
		},
	})
}
