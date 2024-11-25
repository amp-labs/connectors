package api3

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestStarRulePathResolver(t *testing.T) { // nolint:funlen
	t.Parallel()

	type inType struct {
		endpoints []string
		path      string
	}

	tests := []struct {
		name     string
		input    inType
		expected bool
	}{
		{
			name: "No endpoints prohibit all paths",
			input: inType{
				endpoints: []string{},
				path:      "/customer",
			},
			expected: false,
		},
		{
			name: "Exact match",
			input: inType{
				endpoints: []string{"/customer"},
				path:      "/customer",
			},
			expected: true,
		},
		{
			name: "Prefix match",
			input: inType{
				endpoints: []string{"*/search"},
				path:      "/orders/products/search",
			},
			expected: true,
		},
		{
			name: "Suffix match",
			input: inType{
				endpoints: []string{"/v3/*"},
				path:      "/v3/coupons",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolver := newStarRulePathResolver(tt.input.endpoints, func(matched bool) bool {
				return matched
			})
			output := resolver.IsPathMatching(tt.input.path)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}
