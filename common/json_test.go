package common

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestUnmarshalJSON(t *testing.T) { // nolint:funlen
	t.Parallel()

	type Story struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	tests := []struct {
		name        string
		input       []byte
		expected    *Story
		expectedErr error
	}{
		{
			name:  "Nil body produces empty struct",
			input: nil,
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:  "Empty body produces empty struct",
			input: []byte(""),
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid JSON produces marshal error",
			input:       []byte("2359"),
			expected:    nil,
			expectedErr: ErrFailedToUnmarshalBody,
		},
		{
			name:  "Valid JSON values are mapped to struct fields",
			input: []byte(`{"title": "Amazing", "description": "very long story"}`),
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := UnmarshalJSON[Story](&JSONHTTPResponse{
				bodyBytes: tt.input,
			})
			testutils.CheckOutputWithError(t, tt.name, tt.expected, tt.expectedErr, output, err)
		})
	}
}
