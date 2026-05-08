package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// nolint:funlen
func TestConnector_getURL_ModuleCRM(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		baseURL   string // optional; if empty, use defaultBaseURL
		arg       string
		queryArgs []string
		wantURL   string
		wantErr   bool
	}{
		{
			name:    "Read with default baseURL (trailing slash removed)",
			arg:     "objects/contacts/",
			wantURL: "https://api.hubapi.com/crm/v3/objects/contacts",
		},
		{
			name:      "Read with query params (special chars)",
			arg:       "objects/contacts",
			queryArgs: []string{"properties", "email,first name", "archived", "true"},
			wantURL:   "https://api.hubapi.com/crm/v3/objects/contacts?archived=true&properties=email%2Cfirst+name",
		},
		{
			name:      "Error: missing query param value",
			arg:       "objects/contacts",
			queryArgs: []string{"properties"},
			wantErr:   true,
		},
		{
			name:    "BaseURL with extra trailing slash",
			baseURL: "https://api.hubapi.com/crm/", // add an extra slash
			arg:     "objects/contacts/",
			wantURL: "https://api.hubapi.com/crm/v3/objects/contacts",
		},
	}

	for _, ttc := range cases {
		t.Run(ttc.name, func(t *testing.T) {
			t.Parallel()

			c, err := NewConnector(
				common.ConnectorParams{
					Module:              providers.ModuleHubspotCRM,
					AuthenticatedClient: mockutils.NewClient(),
				},
			)

			if ttc.baseURL != "" {
				c.ModuleInfo().BaseURL = ttc.baseURL
			}

			gotURL, err := c.getURL(ttc.arg, ttc.queryArgs...)
			if ttc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotURL != ttc.wantURL {
				t.Errorf("got URL %q, want %q", gotURL, ttc.wantURL)
			}
		})
	}
}
