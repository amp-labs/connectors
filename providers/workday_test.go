package providers

import "testing"

// The provider app is created before any connection exists, so the Workday
// OAuth URLs must resolve from the workspace alone. The tenant and the
// authorization host are only known once a connection supplies them.
func TestWorkdayReadInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vars     []string
		baseURL  string
		authURL  string
		tokenURL string
	}{
		{
			name:     "Resolves without a connection using input defaults",
			vars:     []string{"workspace", "wd2-impl-services1.workday.com"},
			baseURL:  "https://wd2-impl-services1.workday.com",
			authURL:  "https://impl.workday.com/tenant_name/authorize",
			tokenURL: "https://wd2-impl-services1.workday.com/ccx/oauth2/tenant_name/token",
		},
		{
			name: "Connection inputs override the defaults",
			vars: []string{
				"workspace", "wd3-services1.myworkday.com",
				"tenantName", "acme",
				"authHost", "wd3.myworkday.com",
			},
			baseURL:  "https://wd3-services1.myworkday.com",
			authURL:  "https://wd3.myworkday.com/acme/authorize",
			tokenURL: "https://wd3-services1.myworkday.com/ccx/oauth2/acme/token",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := ReadInfo(Workday, createCatalogVars(tt.vars...)...)
			if err != nil {
				t.Fatalf("ReadInfo() error: %v", err)
			}

			if info.BaseURL != tt.baseURL {
				t.Errorf("BaseURL = %q, want %q", info.BaseURL, tt.baseURL)
			}

			if info.Oauth2Opts.AuthURL != tt.authURL {
				t.Errorf("AuthURL = %q, want %q", info.Oauth2Opts.AuthURL, tt.authURL)
			}

			if info.Oauth2Opts.TokenURL != tt.tokenURL {
				t.Errorf("TokenURL = %q, want %q", info.Oauth2Opts.TokenURL, tt.tokenURL)
			}
		})
	}
}
