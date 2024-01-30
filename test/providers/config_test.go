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
		substitutions map[string]string
		expected      map[string]string
		expectedErr   error
	}{
		{
			provider: providers.Salesforce,
			substitutions: map[string]string{
				"subdomain": "example",
			},
			expected: map[string]string{
				"connector_type":     "full",
				"connector_version":  "1.0.0",
				"provider_auth_type": "oauth2",
				"provider_base_url":  "https://example.salesforce.com",
				"provider_version":   "v59.0",
			},
			expectedErr: nil,
		},
		{
			provider: providers.Hubspot,
			substitutions: map[string]string{
				"nonexistentvar": "test",
			},
			expected: map[string]string{
				"connector_type":     "full",
				"connector_version":  "1.0.0",
				"provider_auth_type": "oauth2",
				"provider_base_url":  "https://api.hubapi.com",
			},
			expectedErr: nil,
		},
		{
			provider: providers.LinkedIn,
			substitutions: map[string]string{
				"nonexistentvar": "xyz",
			},
			expected: map[string]string{
				"connector_type":     "basic",
				"connector_version":  "0.1.0",
				"provider_auth_type": "oauth2",
				"provider_base_url":  "https://api.linkedin.com",
				"provider_version":   "2",
			},
			expectedErr: nil,
		},
		{
			provider: providers.Provider("nonexistent"),
			substitutions: map[string]string{
				"subdomain": "test",
			},
			expected:    nil,
			expectedErr: providers.ErrProviderConfigNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.provider), func(t *testing.T) {
			config, err := providers.ReadConfig(tc.provider, tc.substitutions)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil && config != nil {
				for key, value := range config {
					if expected, ok := tc.expected[key]; ok && value != expected {
						t.Errorf("Mismatch for key: %s, expected: %s, got: %s", key, expected, value)
					}
				}
			}
		})
	}
}
