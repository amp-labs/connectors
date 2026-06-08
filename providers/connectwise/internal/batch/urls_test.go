package batch

import (
	"net/url"
	"testing"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWithIdentifiers(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		identifiers  []string
		maxURLLength int
		wantURLs     int
		wantErr      error
	}{
		{
			name:         "Empty identifiers",
			baseURL:      "https://example.com",
			identifiers:  []string{},
			maxURLLength: 100,
			wantURLs:     1,
			wantErr:      nil,
		},
		{
			name:         "Single identifier within limit",
			baseURL:      "https://example.com",
			identifiers:  []string{"123"},
			maxURLLength: 100,
			wantURLs:     1,
			wantErr:      nil,
		},
		{
			name:         "Single identifier exceeding limit",
			baseURL:      "https://example.com",
			identifiers:  []string{"1234567890"},
			maxURLLength: 30, // Very small limit
			wantURLs:     0,
			wantErr:      ErrURLNotEnoughSpace,
		},
		{
			name:         "Multiple identifiers fitting in one URL",
			baseURL:      "https://example.com",
			identifiers:  []string{"1", "2", "3"},
			maxURLLength: 100,
			wantURLs:     1,
			wantErr:      nil,
		},
		{
			name:         "Multiple identifiers split across URLs",
			baseURL:      "https://example.com",
			identifiers:  []string{"123", "456", "789", "_a_", "_b_", "_c_"},
			maxURLLength: 50, // Should split them
			wantURLs:     6,
			wantErr:      nil,
		},
		{
			name:        "Exact fit for multiple IDs",
			baseURL:     "https://example.com",
			identifiers: []string{"123", "456"},
			// base: 19
			// query: len("?conditions=id+in+%28%29") = 24
			// total reserved: 43
			// IDs: "123", "456"
			// sizes: 3 + 3 (comma) + 3 = 9
			// total expected: 43 + 9 = 52
			maxURLLength: 52,
			wantURLs:     1,
			wantErr:      nil,
		},
		{
			name:        "One ID fits, second doesn't",
			baseURL:     "https://example.com",
			identifiers: []string{"123", "456"},
			// reserved: 43
			// first ID: 3
			// comma + second ID: 3 + 3 = 6
			// total if both: 43 + 3 + 6 = 52
			// if only first: 43 + 3 = 46
			maxURLLength: 51,
			wantURLs:     2,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutils.NewCompareResult()

			baseURL, err := urlbuilder.New(tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create baseURL: %v", err)
			}

			got, err := withIdentifiers(baseURL, tt.identifiers, tt.maxURLLength)
			if !result.AssertErr("err", tt.wantErr, err) {
				result.Validate(t, tt.name)
				return
			}

			result.Assert("number of URLs", len(got), tt.wantURLs)

			// Verify URL validity and length
			for _, u := range got {
				result.Assert("procedure miscalculated URL size", len(u.URL), u.estimatedSize)

				if len(u.URL) > tt.maxURLLength {
					result.AddDiff("Generated URL exceeds max length: %d > %d", len(u.URL), tt.maxURLLength)
				}

				if _, err := url.Parse(u.URL); err != nil {
					result.AddDiff("Generated invalid URL: %v", err)
				}
			}

			result.Validate(t, tt.name)
		})
	}
}
