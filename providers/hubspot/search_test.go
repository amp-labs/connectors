package hubspot

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestCheckSearchResultsLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		nextPage    common.NextPageToken
		expectError bool
	}{
		{
			name:        "Empty next page token is valid",
			nextPage:    "",
			expectError: false,
		},
		{
			name:        "Zero offset is valid",
			nextPage:    "0",
			expectError: false,
		},
		{
			name:        "Offset below limit is valid",
			nextPage:    "8999",
			expectError: false,
		},
		{
			name:        "Offset at limit returns error",
			nextPage:    "10000",
			expectError: true,
		},
		{
			name:        "Offset above limit returns error",
			nextPage:    "10001",
			expectError: true,
		},
		{
			name:        "Large offset returns error",
			nextPage:    "99999",
			expectError: true,
		},
		{
			name:        "Non-numeric token is ignored (allowed to proceed)",
			nextPage:    "not-a-number",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := checkSearchResultsLimit(tt.nextPage)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for nextPage=%q, got nil", tt.nextPage)
				}

				if !errors.Is(err, common.ErrResultsLimitExceeded) {
					t.Errorf("expected ErrResultsLimitExceeded, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for nextPage=%q: %v", tt.nextPage, err)
				}
			}
		})
	}
}

func TestSearchResultsLimitConstant(t *testing.T) {
	t.Parallel()

	// Ensure the limit constant matches HubSpot's documented limit
	if searchResultsLimit != 10000 {
		t.Errorf("searchResultsLimit should be 10000, got %d", searchResultsLimit)
	}
}
