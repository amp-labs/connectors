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
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolver := newStarRulePathResolver(tt.input.endpoints)
			output := resolver.IsPathMatching(tt.input.path)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}

func TestPathMatchingStrategy(t *testing.T) { // nolint:funlen
	t.Parallel()

	type inType struct {
		strategy PathMatchingStrategy
		path     string
	}

	tests := []struct {
		name     string
		input    inType
		expected bool
	}{
		{
			name: "Empty strategy matches all paths",
			input: inType{
				strategy: PathMatchingStrategy{
					Strict: false,
				},
				path: "/customer",
			},
			expected: true,
		},
		{
			name: "Path not covered by any matcher prohibit in strict mode",
			input: inType{
				strategy: PathMatchingStrategy{
					AllowPaths: []string{"/products"},
					DenyPaths:  []string{"/logs"},
					Strict:     true,
				},
				path: "/customer",
			},
			expected: false,
		},
		{
			name: "Allow star paths",
			input: inType{
				strategy: PathMatchingStrategy{
					AllowPaths: []string{"/customer/*"},
				},
				path: "/customer/profile",
			},
			expected: true,
		},
		{
			name: "Deny star paths",
			input: inType{
				strategy: PathMatchingStrategy{
					DenyPaths: []string{"/customer/*"},
				},
				path: "/customer/profile",
			},
			expected: false,
		},
		{
			name: "Tie is resolved in favour of Allow",
			input: inType{
				strategy: PathMatchingStrategy{
					AllowPaths:    []string{"/customer/*"},
					DenyPaths:     []string{"*/search"},
					PriorityAllow: true,
				},
				path: "/customer/search",
			},
			expected: true,
		},
		{
			name: "Tie is resolved in favour of Deny",
			input: inType{
				strategy: PathMatchingStrategy{
					AllowPaths:    []string{"/customer/*"},
					DenyPaths:     []string{"*/search"},
					PriorityAllow: false,
				},
				path: "/customer/search",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolver := tt.input.strategy.createResolver()
			output := resolver.IsPathMatching(tt.input.path)
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, nil)
		})
	}
}
