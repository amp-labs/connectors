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
					Delete: true,
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
		provider:    Klaviyo,
		description: "Valid Klaviyo provider config with no substitutions",
		expected: &ProviderInfo{
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
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 "PKCE",
				AuthURL:                   "https://www.klaviyo.com/oauth/authorize",
				TokenURL:                  "https://a.klaviyo.com/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
				TokenMetadataFields: TokenMetadataFields{
					ScopesField: "scope",
				},
			},
			BaseURL: "https://a.klaviyo.com",
		},
		expectedErr: nil,
	},

	{
		provider: Sellsy,
		expected: &ProviderInfo{
			AuthType: Oauth2,
			OauthOpts: OauthOpts{
				GrantType:                 PKCE,
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
			BaseURL: "https://api.dropboxapi.com",
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
