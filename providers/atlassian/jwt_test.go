package atlassian

import (
	"testing"
)

// Test cases for createQueryStringHash.
func TestCreateQueryStringHash(t *testing.T) { // nolint: funlen
	t.Parallel()

	tests := []struct {
		method string
		path   string
		query  map[string][]string

		// These have been computed using atlassian's official jwt library
		expected string
	}{
		{
			method: "GET",
			path:   "/",
			query: map[string][]string{
				"param": {"value"},
			},
			expected: "f0ac26e46a317ce416f2c00803c685edc2aaea1e7212e6b7ef916a6f50ba2dd6",
		},
		{
			method: "post",
			path:   "/api/resource",
			query: map[string][]string{
				"action": {"create"},
				"id":     {"123"},
			},
			expected: "48afa36ccd6c9e8988d19a31ee9d492d964e31b63115ef722d4c8ef94f6dc843",
		},
		{
			method: "GET",
			path:   "/api/resource/",
			query: map[string][]string{
				"param": {"value1", "value2"},
				"jwt":   {"token123"},
			},
			expected: "0f93a4689832e5d5c9ec52448be868b5be2d4bca9d113f99d88f93b67d0d9ed0",
		},
		{
			method: "get",
			path:   "/path/with&special=chars",
			query: map[string][]string{
				"param": {"value with spaces", "special&chars"},
				"jwt":   {"token"},
			},
			expected: "8fac2e94d5cb524f65d99d47dfbacb8701798fb5e4c607a7ddef8f86e2ed8029",
		},
		{
			method: "POST",
			path:   "/api/resource/with special&chars/",
			query: map[string][]string{
				"param1":  {"value1", "value2"},
				"param2":  {"value with spaces", "special&chars"},
				"param&3": {"value=equals", "value?question"},
				"param+4": {"value/slash", "value\\backslash"},
				"jwt":     {"token"}, // Should be filtered out
			},
			expected: "b5bc15c45a7d471c3c1df8d283e02c754b36fbf138468e785a0fccad363d7cd0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()

			result := createQueryStringHash(tt.method, tt.path, tt.query)
			if result != tt.expected {
				t.Errorf("Expected hash %s, got %s", tt.expected, result)
			}
		})
	}
}
