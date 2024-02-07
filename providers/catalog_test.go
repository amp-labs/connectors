package providers

import (
	"errors"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Define test cases
	testCases := []struct {
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
				"subdomain": "example",
				"version":   "-1.0",
			},
			expected: &ProviderInfo{
				Support: ConnectorSupport{
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
				BaseURL: "https://example.salesforce.com",
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
				Support: ConnectorSupport{
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
				Support: ConnectorSupport{
					Read:      false,
					Write:     false,
					BulkWrite: false,
					Subscribe: false,
					Proxy:     false,
				},
				AuthType:  AuthTypeOAuth2,
				OauthOpts: OauthOpts{},
				BaseURL:   "https://api.linkedin.com",
			},
			expectedErr: nil,
		},
		{
			provider:    Provider("nonexistent"),
			description: "Non-existent provider config",
			substitutions: map[string]string{
				"subdomain": "test",
			},
			expected:    nil,
			expectedErr: ErrProviderCatalogNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.provider), func(t *testing.T) {
			config, err := ReadConfig(tc.provider, &tc.substitutions)
			t.Logf("Test case: %s", tc.description)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil && config != nil {
				if config.Support != tc.expected.Support {
					t.Errorf("Expected support: %v, but got: %v", tc.expected.Support, config.Support)
				}

				if config.AuthType != tc.expected.AuthType {
					t.Errorf("Expected auth: %v, but got: %v", tc.expected.AuthType, config.AuthType)
				}

				if config.OauthOpts != tc.expected.OauthOpts {
					t.Errorf("Expected auth options: %v, but got: %v", tc.expected.OauthOpts, config.OauthOpts)
				}

				if config.BaseURL != tc.expected.BaseURL {
					t.Errorf("Expected base URL: %s, but got: %s", tc.expected.BaseURL, config.BaseURL)
				}
			}
		})
	}
}
