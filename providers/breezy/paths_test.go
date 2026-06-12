package breezy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVersionedPathURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		baseURL    string
		objectPath string
		expected   string
	}{
		{
			name:       "Top-level object path",
			baseURL:    "https://api.breezy.hr",
			objectPath: "/companies",
			expected:   "https://api.breezy.hr/v3/companies",
		},
		{
			name:       "Company-scoped object path",
			baseURL:    "https://api.breezy.hr",
			objectPath: "/company/abc123/webhook_endpoints",
			expected:   "https://api.breezy.hr/v3/company/abc123/webhook_endpoints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := buildVersionedPathURL(tt.baseURL, tt.objectPath)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, u.String())
		})
	}
}

func TestResolveObjectPath(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"/company/abc123/webhook_endpoints",
		resolveObjectPath("/company/{company_id}/webhook_endpoints", "abc123"),
	)
	assert.Equal(t, "/companies", resolveObjectPath("/companies", "abc123"))
}
