package providers

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

const Salesforce Provider = "salesforce"

const (
	// ModuleSalesforceCRM
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_what_is_rest_api.htm
	ModuleSalesforceCRM common.ModuleID = "crm"
	// ModuleSalesforceAccountEngagement
	// https://developer.salesforce.com/docs/marketing/pardot/guide/use-cases.html
	ModuleSalesforceAccountEngagement common.ModuleID = "account-engagement"
	// ModuleSalesforceAccountEngagementDemo
	// It is similar to ModuleSalesforceAccountEngagement but targets non-production URL.
	ModuleSalesforceAccountEngagementDemo common.ModuleID = "account-engagement-demo"
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
			ModuleSalesforceAccountEngagement: {
				BaseURL:     "https://pi.pardot.com",
				DisplayName: "Salesforce Marketing Cloud Account Engagement (Pardot)",
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
			ModuleSalesforceAccountEngagementDemo: {
				BaseURL:     "https://pi.demo.pardot.com",
				DisplayName: "Salesforce Demo Marketing Cloud Account Engagement (Pardot)",
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
			Delete:    true,
			Proxy:     true,
			Read:      true,
			Subscribe: true,
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
					// ModuleDependencies specifies which modules REQUIRE this metadata item.
					// Here, it means: "the CRM module depends on/requires the workspace metadata".
					ModuleDependencies: &ModuleDependencies{
						ModuleSalesforceCRM:                   {},
						ModuleSalesforceAccountEngagement:     {},
						ModuleSalesforceAccountEngagementDemo: {},
					},
				},
				// {
				// 	DisplayName: "Business Unit ID",
				// 	DocsURL:     "https://help.salesforce.com/s/articleView?id=000381973&type=1",
				// 	ModuleDependencies: &ModuleDependencies{
				// 		ModuleSalesforceAccountEngagement:     {},
				// 		ModuleSalesforceAccountEngagementDemo: {},
				// 	},
				// 	Name:   "businessUnitId",
				// 	Prompt: "Business Unit ID is found in Business Unit Setup within Salesforce Setup or Marketing Setup.",
				// },
			},
		},
	})
}
