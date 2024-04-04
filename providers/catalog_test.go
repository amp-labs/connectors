package providers

import (
	"errors"
	"testing"
)

// All test cases.
var testCases = []struct { // nolint
	provider      Provider
	description   string
	substitutions map[string]string
	expected      *ProviderInfo
	expectedErr   error
}{
	{
		provider:    Salesforce,
		description: "Salesforce provider config with valid & invalid substitutions",
		substitutions: map[string]string{
			"workspace": "example",
			"version":   "-1.0",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  true,
				Write: true,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: true,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     true,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 AuthorizationCode,
				AuthURL:                   "https://example.my.salesforce.com/services/oauth2/authorize",
				TokenURL:                  "https://example.my.salesforce.com/services/oauth2/token",
				ExplicitWorkspaceRequired: true,
				ExplicitScopesRequired:    false,
				TokenMetadataFields: TokenMetadataFields{
					ConsumerRefField:  "id",
					WorkspaceRefField: "instance_url",
					ScopesField:       "scope",
				},
			},
			BaseURL: "https://example.my.salesforce.com",
			ProviderOpts: ProviderOpts{
				"restApiUrl": "https://example.my.salesforce.com/services/data/v59.0",
				"domain":     "example.my.salesforce.com",
			},
		},
		expectedErr: nil,
	},
	{
		provider:    Hubspot,
		description: "Valid hubspot provider config with non-existent substitutions",
		substitutions: map[string]string{
			"nonexistentvar": "test",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  true,
				Write: true,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     true,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 AuthorizationCode,
				AuthURL:                   "https://app.hubspot.com/oauth/authorize",
				TokenURL:                  "https://api.hubapi.com/oauth/v1/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.hubapi.com",
		},
		expectedErr: nil,
	},
	{
		provider:    LinkedIn,
		description: "Valid LinkedIn provider config with non-existent substitutions",
		substitutions: map[string]string{
			"nonexistentvar": "xyz",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 AuthorizationCode,
				AuthURL:                   "https://www.linkedin.com/oauth/v2/authorization",
				TokenURL:                  "https://www.linkedin.com/oauth/v2/accessToken",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField: "scope",
				},
			},
			BaseURL: "https://api.linkedin.com",
		},
		expectedErr: nil,
	},
	{
		provider:    "nonexistent",
		description: "Non-existent provider config",
		substitutions: map[string]string{
			"workspace": "test",
		},
		expected:    nil,
		expectedErr: ErrProviderCatalogNotFound,
	},
	{
		provider:    Salesloft,
		description: "Valid SalesLoft provider config with non-existent substitutions",
		substitutions: map[string]string{
			"nonexistentvar": "abc",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://accounts.salesloft.com/oauth/authorize",
				TokenURL:                  "https://accounts.salesloft.com/oauth/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.salesloft.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Outreach,
		description: "Valid Outreach provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://api.outreach.io/oauth/authorize",
				TokenURL:                  "https://api.outreach.io/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.outreach.io",
		},
		expectedErr: nil,
	},

	{
		provider: Pipedrive,
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://oauth.pipedrive.com/oauth/authorize",
				TokenURL:                  "https://oauth.pipedrive.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
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
			BaseURL: "https://api.pipedrive.com",
		},
		expectedErr: nil,
	},

	{
		provider: Capsule,
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://api.capsulecrm.com/oauth/authorise",
				TokenURL:                  "https://api.capsulecrm.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
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
			BaseURL: "https://api.capsulecrm.com/api",
		},
		expectedErr: nil,
	},

	{
		provider:    Copper,
		description: "Valid Copper provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.copper.com/oauth/authorize",
				TokenURL:                  "https://app.copper.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.copper.com/developer_api",
		},
		expectedErr: nil,
	},

	{
		provider:    ZohoCRM,
		description: "Valid ZohoCRM provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
				TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://www.zohoapis.com",
		},
		expectedErr: nil,
	},

	{
		provider: Sellsy,
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://login.sellsy.com/oauth2/authorization",
				TokenURL:                  "https://login.sellsy.com/oauth2/access-tokens",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
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
			BaseURL: "https://api.sellsy.com",
		},
		expectedErr: nil,
	},

	{
		provider:    Attio,
		description: "Valid Attio provider config with non-existent substitutions",
		substitutions: map[string]string{
			"nonexistentvar": "abc",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.attio.com/authorize",
				TokenURL:                  "https://app.attio.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.attio.com/api",
		},
		expectedErr: nil,
	},

	{
		provider:    Close,
		description: "Valid Close provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,

				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.close.com/oauth2/authorize",
				TokenURL:                  "https://api.close.com/oauth2/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.close.com/api",
		},
		expectedErr: nil,
	},

	{
		provider:    Keap,
		description: "Valid Keap provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,

				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://accounts.infusionsoft.com/app/oauth/authorize",
				TokenURL:                  "https://api.infusionsoft.com/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.infusionsoft.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Asana,
		description: "Valid Asana provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.asana.com/-/oauth_authorize",
				TokenURL:                  "https://app.asana.com/-/oauth_token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://app.asana.com/api",
		},
		expectedErr: nil,
	},
	{
		provider: Dropbox,
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://www.dropbox.com/oauth2/authorize",
				TokenURL:                  "https://api.dropboxapi.com/oauth2/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
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
			BaseURL: "https://api.dropboxapi.com/2/",
		},
		expectedErr: nil,
	},
	{
		provider:    Notion,
		description: "Valid Notion provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     true,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://api.notion.com/v1/oauth/authorize",
				TokenURL:                  "https://api.notion.com/v1/oauth/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
				TokenMetadataFields: TokenMetadataFields{
					WorkspaceRefField: "workspace_id",
					ConsumerRefField:  "owner.user.id",
				},
			},
			BaseURL: "https://api.notion.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Gong,
		description: "Gong provider config with valid substitutions",
		substitutions: map[string]string{
			"workspace": "testing",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.gong.io/oauth2/authorize",
				TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
				ExplicitWorkspaceRequired: false,
				ExplicitScopesRequired:    true,
			},
			BaseURL: "https://testing.api.gong.io",
		},
		expectedErr: nil,
	},
	{
		provider:    Zoom,
		description: "Zoom provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://zoom.us/oauth/authorize",
				TokenURL:                  "https://zoom.us/oauth/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.zoom.us",
		},
		expectedErr: nil,
	},
	{
		provider:    Intercom,
		description: "Valid Intercom provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.intercom.com/oauth",
				TokenURL:                  "https://api.intercom.io/auth/eagle/token",
				ExplicitWorkspaceRequired: false,
				ExplicitScopesRequired:    false,
			},
			BaseURL: "https://api.intercom.io",
		},
		expectedErr: nil,
	},
	{
		provider: DocuSign,
		substitutions: map[string]string{
			"workspace": "example",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://account.docusign.com/oauth/auth",
				TokenURL:                  "https://account.docusign.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: true,
			},
			BaseURL: "https://example.docusign.net",
		},
		expectedErr: nil,
	},
	{
		provider: DocuSignDeveloper,
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://account-d.docusign.com/oauth/auth",
				TokenURL:                  "https://account-d.docusign.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://demo.docusign.net",
		},
		expectedErr: nil,
	},
	{
		provider:    Calendly,
		description: "Calendly provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://auth.calendly.com/oauth/authorize",
				TokenURL:                  "https://auth.calendly.com/oauth/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.calendly.com",
		},
		expectedErr: nil,
	},
	{
		provider:    GetResponse,
		description: "GetResponse provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://app.getresponse.com/oauth2_authorize.html",
				TokenURL:                  "https://api.getresponse.com/v3/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.getresponse.com",
		},
		expectedErr: nil,
	},
	{
		provider:    AWeber,
		description: "Valid AWeber provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://auth.aweber.com/oauth2/authorize",
				TokenURL:                  "https://auth.aweber.com/oauth2/token",
				ExplicitWorkspaceRequired: false,
				ExplicitScopesRequired:    true,
			},
			BaseURL: "https://api.aweber.com",
		},
		expectedErr: nil,
	},
	{
		provider:    MicrosoftDynamics365Sales,
		description: "MS Dynamics 365 Sales provider config with valid substitutions",
		substitutions: map[string]string{
			"workspace": "testing",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL:                  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: true,
			},
			BaseURL: "https://testing.api.crm.dynamics.com",
		},
		expectedErr: nil,
	},
	{
		provider:    ConstantContact,
		description: "Valid ConstantContact provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://authz.constantcontact.com/oauth2/default/v1/authorize",
				TokenURL:                  "https://authz.constantcontact.com/oauth2/default/v1/token",
				ExplicitWorkspaceRequired: false,
				ExplicitScopesRequired:    true,
			},
			BaseURL: "https://api.cc.email",
		},
		expectedErr: nil,
	},
	{
		provider:    MicrosoftDynamics365BusinessCentral,
		description: "Dynamics 365 Business Central provider config with substitutions",
		substitutions: map[string]string{
			"workspace": "tenantID",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://login.microsoftonline.com/tenantID/oauth2/v2.0/authorize",
				TokenURL:                  "https://login.microsoftonline.com/tenantID/oauth2/v2.0/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: true,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField: "scope",
				},
			},
			BaseURL: "https://api.businesscentral.dynamics.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Gainsight,
		description: "Gainsight config with substitutions",
		substitutions: map[string]string{
			"workspace": "company",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://company.gainsightcloud.com/v1/authorize",
				TokenURL:                  "https://company.gainsightcloud.com/v1/users/oauth/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: true,
			},
			BaseURL: "https://company.gainsightcloud.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Box,
		description: "Box config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://account.box.com/api/oauth2/authorize",
				TokenURL:                  "https://api.box.com/oauth2/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
			},
			BaseURL: "https://api.box.com",
		},
		expectedErr: nil,
	},
	{
		provider:    GoogleCalendar,
		description: "Google Calendar provider config with no substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL:                  "https://oauth2.googleapis.com/token",
				ExplicitWorkspaceRequired: false,
				ExplicitScopesRequired:    true,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField: "scope",
				},
			},
			BaseURL: "https://www.googleapis.com/calendar",
		},
		expectedErr: nil,
	},
	{
		provider:    ZendeskSupport,
		description: "Zendesk Support provider config with valid substitutions",
		substitutions: map[string]string{
			"workspace": "testing",
		},
		expected: &ProviderInfo{
			Support: Support{
				Read:  false,
				Write: false,
				BulkWrite: BulkWriteSupport{
					Insert: false,
					Update: false,
					Upsert: false,
					Delete: false,
				},
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://testing.zendesk.com/oauth/authorizations/new",
				TokenURL:                  "https://testing.zendesk.com/oauth/tokens",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: true,
			},
			BaseURL: "https://testing.zendesk.com",
		},
		expectedErr: nil,
	},
	{
		provider:    ZendeskChat,
		description: "Valid ZendeskChat provider config with substitutions",
		substitutions: map[string]string{
			"workspace": "test",
		},
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://www.zopim.com/oauth2/authorizations/new?subdomain=test",
				TokenURL:                  "https://www.zopim.com/oauth2/token",
				ExplicitScopesRequired:    true,
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
			BaseURL: "https://www.zopim.com",
		},
		expectedErr: nil,
	},
	{
		provider:    WordPress,
		description: "Valid WordPress provider config with no substitutions",
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
				TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
				ExplicitScopesRequired:    false,
				ExplicitWorkspaceRequired: false,
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
			BaseURL: "https://public-api.wordpress.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Airtable,
		description: "Valid Airtable provider config with no substitutions",
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 PKCE,
				AuthURL:                   "https://airtable.com/oauth2/v1/authorize",
				TokenURL:                  "https://airtable.com/oauth2/v1/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField: "scope",
				},
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
			BaseURL: "https://api.airtable.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Slack,
		description: "Valid Slack provider config with non-existent substitutions",
		expected: &ProviderInfo{
			Support: Support{
				Read:      false,
				Write:     false,
				BulkWrite: false,
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 AuthorizationCode,
				AuthURL:                   "https://slack.com/oauth/v2/authorize",
				TokenURL:                  "https://slack.com/api/oauth.v2.access",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: true,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField:       "scope",
					WorkspaceRefField: "workspace_name",
				},
			},
			BaseURL: "https://api.slack.com",
		},
		expectedErr: nil,
	},
}

func TestReadInfo(t *testing.T) { // nolint
	t.Parallel()

	for _, tc := range testCases {
		tc := tc // nolint:varnamelen

		t.Run(tc.provider, func(t *testing.T) {
			t.Parallel()

			config, err := ReadInfo(tc.provider, &tc.substitutions)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("[%s] Expected error: %v, but got: %v", tc.description, tc.expectedErr, err)
			}

			if tc.expectedErr == nil && config != nil { // nolint
				if config.Support != tc.expected.Support {
					t.Errorf("[%s] Expected support: %v, but got: %v", tc.description, tc.expected.Support, config.Support)
				}

				if config.AuthType != tc.expected.AuthType {
					t.Errorf("[%s] Expected auth: %v, but got: %v", tc.description, tc.expected.AuthType, config.AuthType)
				}

				if config.OauthOpts != tc.expected.OauthOpts {
					t.Errorf("[%s] Expected auth options: %v, but got: %v", tc.description, tc.expected.OauthOpts, config.OauthOpts)
				}

				if config.BaseURL != tc.expected.BaseURL {
					t.Errorf("[%s] Expected base URL: %s, but got: %s", tc.description, tc.expected.BaseURL, config.BaseURL)
				}

				if config.ProviderOpts != nil {
					for k, v := range config.ProviderOpts {
						candidateValue, ok := tc.expected.GetOption(k)
						if !ok {
							t.Errorf("[%s] Unexpected option: %s", tc.description, k)
						}

						if v != candidateValue {
							t.Errorf("[%s] Expected option %s: %s, but got: %s", tc.description, k, candidateValue, v)
						}
					}
				}
			}
		})
	}
}
