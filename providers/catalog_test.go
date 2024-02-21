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
				Read:      true,
				Write:     true,
				BulkWrite: true,
				Subscribe: false,
				Proxy:     true,
			},
			AuthType: AuthTypeOAuth2,
			OauthOpts: OauthOpts{
				AuthURL:  "https://example.my.salesforce.com/services/oauth2/authorize",
				TokenURL: "https://example.my.salesforce.com/services/oauth2/token",
			},
			BaseURL: "https://example.my.salesforce.com",
			ProviderOpts: map[string]string{
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
				Read:      true,
				Write:     true,
				BulkWrite: false,
				Subscribe: false,
				Proxy:     true,
			},
			AuthType: AuthTypeOAuth2,
			OauthOpts: OauthOpts{
				AuthURL:  "https://app.hubspot.com/oauth/authorize",
				TokenURL: "https://api.hubapi.com/oauth/v1/token",
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
				Read:      false,
				Write:     false,
				BulkWrite: false,
				Subscribe: false,
				Proxy:     false,
			},
			AuthType: AuthTypeOAuth2,
			OauthOpts: OauthOpts{
				AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
				TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
			},
			BaseURL: "https://api.linkedin.com",
		},
		expectedErr: nil,
	},
	{
		provider:    Provider("nonexistent"),
		description: "Non-existent provider config",
		substitutions: map[string]string{
			"workspace": "test",
		},
		expected:    nil,
		expectedErr: ErrProviderCatalogNotFound,
	},
}

func TestReadConfig(t *testing.T) { // nolint
	t.Parallel()

	for _, tc := range testCases {
		tc := tc // nolint:varnamelen

		t.Run(string(tc.provider), func(t *testing.T) {
			t.Parallel()

			config, err := ReadConfig(tc.provider, &tc.substitutions)

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

				for k, v := range config.ProviderOpts {
					if tc.expected.ProviderOpts[k] != v {
						t.Errorf("[%s] Expected provider option %s: %s, but got: %s", tc.description, k, tc.expected.ProviderOpts[k], v)
					}
				}
			}
		})
	}
}
