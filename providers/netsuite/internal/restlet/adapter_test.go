package restlet

import (
	"strings"
	"testing"
)

func TestBuildRestletURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		baseURL     string
		scriptURL   string
		scriptId    string
		deployId    string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:      "scriptURL alone works",
			baseURL:   "https://1234567.restletsuite.netsuite.com",
			scriptURL: "/app/site/hosting/restlet.nl?script=3277&deploy=1",
			scriptId:  "",
			deployId:  "",
			want:      "https://1234567.restletsuite.netsuite.com/app/site/hosting/restlet.nl?script=3277&deploy=1",
		},
		{
			name:      "scriptId and deployId alone works",
			baseURL:   "https://1234567.restletsuite.netsuite.com",
			scriptURL: "",
			scriptId:  "21",
			deployId:  "7",
			want:      "https://1234567.restletsuite.netsuite.com/app/site/hosting/restlet.nl?deploy=7&script=21",
		},
		{
			name:      "scriptURL takes precedence over scriptId and deployId",
			baseURL:   "https://1234567.restletsuite.netsuite.com",
			scriptURL: "/app/site/hosting/restlet.nl?script=3277&deploy=1",
			scriptId:  "21",
			deployId:  "7",
			want:      "https://1234567.restletsuite.netsuite.com/app/site/hosting/restlet.nl?script=3277&deploy=1",
		},
		{
			name:      "leading slash can be skipped for scriptURL",
			baseURL:   "https://1234567.restletsuite.netsuite.com",
			scriptURL: "app/site/hosting/restlet.nl?script=3277&deploy=1",
			scriptId:  "",
			deployId:  "",
			want:      "https://1234567.restletsuite.netsuite.com/app/site/hosting/restlet.nl?script=3277&deploy=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildRestletURL(tt.baseURL, tt.scriptURL, tt.scriptId, tt.deployId)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("buildRestletURL() error = nil, want error")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("buildRestletURL() err = %v, want substring %q", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildRestletURL() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildRestletURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
