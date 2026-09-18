package hubspot

import "testing"

func TestAPIDomainForHublet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		hublet string
		want   string
	}{
		{name: "na1 keeps the unsuffixed host", hublet: "na1", want: "api.hubapi.com"},
		{name: "eu1 addresses its own hublet", hublet: "eu1", want: "api-eu1.hubapi.com"},
		{name: "case and padding are tolerated", hublet: " EU1 ", want: "api-eu1.hubapi.com"},
		{name: "future hublets follow the same shape", hublet: "ap2", want: "api-ap2.hubapi.com"},
		// A connection that predates the field, or a provider response we don't
		// recognize, must land on today's behavior rather than a guessed host.
		{name: "missing location falls back", hublet: "", want: "api.hubapi.com"},
		{name: "unrecognized location falls back", hublet: "somewhere-else", want: "api.hubapi.com"},
		{name: "injection attempt falls back", hublet: "eu1.evil.com", want: "api.hubapi.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := APIDomainForHublet(tt.hublet); got != tt.want {
				t.Fatalf("APIDomainForHublet(%q) = %q, want %q", tt.hublet, got, tt.want)
			}
		})
	}
}
