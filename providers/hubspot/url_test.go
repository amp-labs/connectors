package hubspot

import (
	"testing"

	"github.com/amp-labs/connectors/providers"
)

func TestConnector_getURL_ModuleCRM(t *testing.T) {
	providerInfo, err := providers.ReadInfo(providers.Hubspot)
	if err != nil {
		t.Fatalf("failed to get providerInfo: %v", err)
	}

	moduleInfo := providerInfo.ReadModuleInfo(providers.ModuleHubspotCRM)
	defaultBaseURL := providerInfo.BaseURL

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
			wantURL: defaultBaseURL + "/crm/v3/objects/contacts",
		},
		{
			name:      "Read with query params (special chars)",
			arg:       "objects/contacts",
			queryArgs: []string{"properties", "email,first name", "archived", "true"},
			wantURL:   defaultBaseURL + "/crm/v3/objects/contacts?archived=true&properties=email%2Cfirst+name",
		},
		{
			name:      "Error: missing query param value",
			arg:       "objects/contacts",
			queryArgs: []string{"properties"},
			wantErr:   true,
		},
		{
			name:    "BaseURL with extra trailing slash",
			baseURL: defaultBaseURL + "/", // add an extra slash
			arg:     "objects/contacts/",
			wantURL: defaultBaseURL + "/crm/v3/objects/contacts",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pi := *providerInfo // copy to avoid race
			mi := *moduleInfo
			if tc.baseURL != "" {
				pi.BaseURL = tc.baseURL
			} else {
				pi.BaseURL = defaultBaseURL
			}

			c := &Connector{
				providerInfo: &pi,
				moduleInfo:   &mi,
			}

			gotURL, err := c.getURL(tc.arg, tc.queryArgs...)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tc.wantURL {
				t.Errorf("got URL %q, want %q", gotURL, tc.wantURL)
			}
		})
	}
}
