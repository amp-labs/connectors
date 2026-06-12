package breezy

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

type versionedPathInput struct {
	BaseURL    string
	ObjectPath string
}

func TestBuildVersionedPathURL(t *testing.T) {
	t.Parallel()

	tests := []testroutines.TestCase[versionedPathInput, string]{
		{
			Name:   "Top-level object path",
			Server: mockserver.Dummy(),
			Input: versionedPathInput{
				BaseURL:    "https://api.breezy.hr",
				ObjectPath: "/companies",
			},
			Expected: "https://api.breezy.hr/v3/companies",
		},
		{
			Name:   "Company-scoped object path",
			Server: mockserver.Dummy(),
			Input: versionedPathInput{
				BaseURL:    "https://api.breezy.hr",
				ObjectPath: "/company/abc123/webhook_endpoints",
			},
			Expected: "https://api.breezy.hr/v3/company/abc123/webhook_endpoints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(tt.Close)

			input := tt.PrepareInput()
			url, err := buildVersionedPathURL(input.BaseURL, input.ObjectPath)

			var got string
			if url != nil {
				got = url.String()
			}

			tt.Validate(t, err, got)
		})
	}
}

type resolvePathInput struct {
	ObjectPath string
	CompanyID  string
}

func TestResolveObjectPath(t *testing.T) {
	t.Parallel()

	tests := []testroutines.TestCase[resolvePathInput, string]{
		{
			Name:   "Substitutes company id placeholder",
			Server: mockserver.Dummy(),
			Input: resolvePathInput{
				ObjectPath: "/company/{company_id}/webhook_endpoints",
				CompanyID:  "abc123",
			},
			Expected: "/company/abc123/webhook_endpoints",
		},
		{
			Name:   "Leaves path unchanged when no placeholder",
			Server: mockserver.Dummy(),
			Input: resolvePathInput{
				ObjectPath: "/companies",
				CompanyID:  "abc123",
			},
			Expected: "/companies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(tt.Close)

			input := tt.PrepareInput()
			got := resolveObjectPath(input.ObjectPath, input.CompanyID)

			tt.Validate(t, nil, got)
		})
	}
}
