// nolint:revive
package common

import (
	"bytes"
	"io"
	"net/http"
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
		contentType string
		expected    *Story
		expectedErr error
	}{
		{
			name:        "Nil body produces empty struct",
			input:       nil,
			contentType: "application/json",
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:        "Empty body produces empty struct",
			input:       []byte(""),
			contentType: "application/json",
			expected: &Story{
				Title:       "",
				Description: "",
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid JSON produces marshal error",
			input:       []byte("2359"),
			contentType: "application/json",
			expected:    nil,
			expectedErr: ErrFailedToUnmarshalBody,
		},
		{
			name:        "Valid JSON values are mapped to struct fields",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Valid JSON with application/vnd.api+json content type",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/vnd.api+json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Valid JSON with application/schema+json content type",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/schema+json",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid content type text/plain",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "text/plain",
			expected:    nil,
			expectedErr: ErrNotJSON,
		},
		{
			name:        "Invalid content type application/xml",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "application/xml",
			expected:    nil,
			expectedErr: ErrNotJSON,
		},
		{
			name:        "Empty content type, assume application/json and try to unmarshal",
			input:       []byte(`{"title": "Amazing", "description": "very long story"}`),
			contentType: "",
			expected: &Story{
				Title:       "Amazing",
				Description: "very long story",
			},
			expectedErr: nil,
		},
	}

	for _, ttc := range tests {
		// nolint:varnamelen
		t.Run(ttc.name, func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				Header: http.Header{
					"Content-Type": []string{ttc.contentType},
				},
				Body: io.NopCloser(bytes.NewReader(ttc.input)),
			}

			output, err := ParseJSONResponse(resp, ttc.input)
			if err != nil {
				testutils.CheckErrors(t, ttc.name, []error{ttc.expectedErr}, err)

				return
			}

			story, outErr := UnmarshalJSON[Story](output)
			testutils.CheckOutputWithError(t, ttc.name, ttc.expected, ttc.expectedErr, story, outErr)
		})
	}
}
