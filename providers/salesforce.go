package providers

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
)

const Salesforce Provider = "salesforce"

const (
	// ModuleSalesforceStandard
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_what_is_rest_api.htm
	ModuleSalesforceStandard common.ModuleID = "standard"
	// ModuleSalesforceAccountEngagement
	// https://developer.salesforce.com/docs/marketing/pardot/guide/use-cases.html
	ModuleSalesforceAccountEngagement common.ModuleID = "account-engagement"
)

func init() { // nolint:funlen
	// Salesforce configuration
	SetInfo(Salesforce, ProviderInfo{
		DisplayName: "Salesforce",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.my.salesforce.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.workspace}}.my.salesforce.com/services/oauth2/userinfo",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.my.salesforce.com/services/oauth2/authorize",
			TokenURL:                  "https://{{.workspace}}.my.salesforce.com/services/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "id",
				WorkspaceRefField: "instance_url",
				ScopesField:       "scope",
			},
		},
		DefaultModule: ModuleSalesforceStandard,
		Modules: &Modules{
			ModuleSalesforceStandard: {
				BaseURL:     "https://{{.workspace}}.my.salesforce.com",
				DisplayName: "Standard Salesforce Platform",
				Support: Support{
					BulkWrite: BulkWriteSupport{
						Insert: false,
						Update: false,
						Upsert: true,
						Delete: true,
					},
					Proxy:     true,
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleSalesforceAccountEngagement: {
				// Workspace can either be empty or ".demo".
				BaseURL:     "https://pi{{.workspace}}.pardot.com",
				DisplayName: "Account Engagement (Pardot)",
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
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: true,
				Delete: true,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
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
				},
			},
		},
	})
}
