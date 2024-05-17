package urlbuilder

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewURL(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:  "URL without query params",
			input: "http://video.google.co.uk:80/videoplay",
		},
		{
			name:  "URL with one query",
			input: "foo://example.com:8042/over/there?name=ferret",
		},
		{
			name:  "URL with multiple queries and fragment",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
		},
		{
			name:  "URL with fragment",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en#00h02m30s",
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := New(tt.input)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
			}

			output := u.String()

			if !reflect.DeepEqual(output, tt.input) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.input, output)
			}
		})
	}
}

func TestWithQueryParam(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		modifier func(*URL)
		expected string
	}{
		{
			name:  "Add one query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParam("compact", "True")
			},
			expected: "https://video.google.co.uk:80/videoplay?compact=True&docid=-7246927612831078230&hl=en",
		},
		{
			name:  "Add list query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParamList("select", []string{"name", "address"})
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en&select=name&select=address",
		},
		{
			name:  "Replace query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParam("hl", "fr")
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=fr",
		},
		{
			name:  "Remove query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.RemoveQueryParam("docid")
			},
			expected: "https://video.google.co.uk:80/videoplay?hl=en",
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := New(tt.input)
			if err != nil {
				t.Fatalf("bad test (%v)", tt.name)
			}

			// apply modifications from test scenario
			tt.modifier(u)
			output := u.String()

			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}
