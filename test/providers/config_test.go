package providers

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/providers"
)

func TestReadConfig(t *testing.T) {
	// Define test cases
	testCases := []struct {
		provider      providers.Provider
		description   string
		substitutions map[string]string
		expected      *providers.ProviderConfig
		expectedErr   error
	}{
		{
			provider:    providers.Salesforce,
			description: "Salesforce provider config with valid & invalid substitutions",
			substitutions: map[string]string{
				"subdomain": "example",
				"version":   "-1.0",
			},
			expected: &providers.ProviderConfig{
				Support: providers.ConnectorSupport{
					Read:      true,
					Write:     true,
					BulkWrite: true,
					Subscribe: false,
					Proxy:     true,
				},
				Auth: providers.Auth{
					Type:     providers.AuthTypeOAuth2,
					AuthURL:  "https://example.my.salesforce.com/services/oauth2/authorize",
					TokenURL: "https://example.my.salesforce.com/services/oauth2/token",
				},
				BaseURL: "https://example.salesforce.com",
			},
			expectedErr: nil,
		},
		{
			provider:    providers.Hubspot,
			description: "Valid hubspot provider config with non-existent substitutions",
			substitutions: map[string]string{
				"nonexistentvar": "test",
			},
			expected: &providers.ProviderConfig{
				Support: providers.ConnectorSupport{
					Read:      true,
					Write:     true,
					BulkWrite: false,
					Subscribe: false,
					Proxy:     true,
				},
				Auth: providers.Auth{
					Type:     providers.AuthTypeOAuth2,
					AuthURL:  "https://app.hubspot.com/oauth/authorize",
					TokenURL: "https://api.hubapi.com/oauth/v1/token",
				},
				BaseURL: "https://api.hubapi.com",
			},
			expectedErr: nil,
		},
		{
			provider:    providers.LinkedIn,
			description: "Valid LinkedIn provider config with non-existent substitutions",
			substitutions: map[string]string{
				"nonexistentvar": "xyz",
			},
			expected: &providers.ProviderConfig{
				Support: providers.ConnectorSupport{
					Read:      false,
					Write:     false,
					BulkWrite: false,
					Subscribe: false,
					Proxy:     false,
				},
				Auth: providers.Auth{
					Type: providers.AuthTypeOAuth2,
				},
				BaseURL: "https://api.linkedin.com",
			},
			expectedErr: nil,
		},
		{
			provider:    providers.Provider("nonexistent"),
			description: "Non-existent provider config",
			substitutions: map[string]string{
				"subdomain": "test",
			},
			expected:    nil,
			expectedErr: providers.ErrProviderConfigNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.provider), func(t *testing.T) {
			config, err := providers.ReadConfig(tc.provider, &tc.substitutions)
			t.Logf("Test case: %s", tc.description)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil && config != nil {
				if config.Support != tc.expected.Support {
					t.Errorf("Expected support: %v, but got: %v", tc.expected.Support, config.Support)
				}

				if config.Auth != tc.expected.Auth {
					t.Errorf("Expected auth: %v, but got: %v", tc.expected.Auth, config.Auth)
				}

				if config.BaseURL != tc.expected.BaseURL {
					t.Errorf("Expected base URL: %s, but got: %s", tc.expected.BaseURL, config.BaseURL)
				}
			}
		})
	}
}
