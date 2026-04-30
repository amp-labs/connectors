package procore

import "testing"

func TestNextPageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		linkHeader string
		want       string
	}{
		{
			name:       "empty header returns empty string",
			linkHeader: "",
			want:       "",
		},
		{
			name:       "header without rel=next returns empty string",
			linkHeader: `<https://api.procore.com/rest/v1.0/companies/123/projects?page=10>; rel="last"`,
			want:       "",
		},
		{
			name:       "extracts the rel=next URL verbatim",
			linkHeader: `<https://api.procore.com/rest/v1.0/companies/123/projects?page=2&per_page=100>; rel="next"`,
			want:       "https://api.procore.com/rest/v1.0/companies/123/projects?page=2&per_page=100",
		},
		{
			name: "picks rel=next when multiple rels are present",
			linkHeader: `<https://api.procore.com/rest/v1.0/companies/123/projects?page=2&per_page=100>; rel="next", ` +
				`<https://api.procore.com/rest/v1.0/companies/123/projects?page=10&per_page=100>; rel="last"`,
			want: "https://api.procore.com/rest/v1.0/companies/123/projects?page=2&per_page=100",
		},
		{
			name: "preserves filters[updated_at] in the URL",
			linkHeader: `<https://api.procore.com/rest/v1.0/companies/4283186/checklist/list_templates` +
				`?filters%5Bupdated_at%5D=2025-10-01T00%3A00%3A00Z...2026-04-17T23%3A59%3A59Z` +
				`&page=2&per_page=2>; rel="next"`,
			want: "https://api.procore.com/rest/v1.0/companies/4283186/checklist/list_templates" +
				"?filters%5Bupdated_at%5D=2025-10-01T00%3A00%3A00Z...2026-04-17T23%3A59%3A59Z" +
				"&page=2&per_page=2",
		},
		{
			name:       "missing angle brackets returns empty string",
			linkHeader: `https://api.procore.com/rest/v1.0/companies/123/projects?page=2; rel="next"`,
			want:       "",
		},
		{
			name:       "malformed URL returns empty string",
			linkHeader: `<http://[::1:bad-url>; rel="next"`,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := nextPageURL(tt.linkHeader)
			if got != tt.want {
				t.Errorf("nextPageURL(%q) = %q, want %q", tt.linkHeader, got, tt.want)
			}
		})
	}
}
