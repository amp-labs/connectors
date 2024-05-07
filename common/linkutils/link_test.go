package linkutils

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestQueryValuesToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    URL
		expected string
	}{
		{
			name: "No params no query string",
			input: URL{
				queryParams: nil,
			},
			expected: "",
		},
		{
			name:     "One parameter",
			input:    createURLWithQuery("$select", []string{"cat"}),
			expected: "?$select=cat",
		},
		{
			name:     "Many parameters",
			input:    createURLWithQuery("$select", []string{"cat", "dog", "parrot", "hamster"}),
			expected: "?$select=cat,dog,parrot,hamster",
		},
		{
			name:     "OData parameters with @ symbol",
			input:    createURLWithQuery("$select", []string{"cat", "@odata.dog", "parrot"}),
			expected: "?$select=cat,@odata.dog,parrot",
		},
	}

	for _, tt := range tests {
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := tt.input.queryValuesToString()
			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

func createURLWithQuery(key string, values []string) URL {
	value := strings.Join(values, ",")

	return URL{
		queryParams: map[string][]string{key: {value}},
	}
}

func TestNewURL(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    *URL
		expectedErr error
	}{
		{
			name:  "URL without query params",
			input: "http://video.google.co.uk:80/videoplay",
			expected: &URL{
				baseURL:     "http://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{},
			},
		},
		{
			name:  "URL with one query",
			input: "foo://example.com:8042/over/there?name=ferret",
			expected: &URL{
				baseURL: "foo://example.com:8042/over/there",
				queryParams: map[string][]string{
					"name": {"ferret"},
				},
			},
		},
		{
			name:  "URL with multiple queries and fragment",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"docid": {"-7246927612831078230"},
					"hl":    {"en"},
				},
			},
		},
		{
			name:  "URL with fragment",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en#00h02m30s",
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"docid": {"-7246927612831078230"},
					"hl":    {"en"},
				},
				fragment: "00h02m30s",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewURL(tt.input)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expectedErr, err)
			}

			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
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
		expected *URL
	}{
		{
			name:  "Add one query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParam("compact", "True")
			},
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"docid":   {"-7246927612831078230"},
					"hl":      {"en"},
					"compact": {"True"},
				},
			},
		},
		{
			name:  "Add list query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParamList("select", []string{"name", "address"})
			},
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"docid":  {"-7246927612831078230"},
					"hl":     {"en"},
					"select": {"name", "address"},
				},
			},
		},
		{
			name:  "Replace query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithQueryParam("hl", "fr")
			},
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"docid": {"-7246927612831078230"},
					"hl":    {"fr"},
				},
			},
		},
		{
			name:  "Remove query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.RemoveQueryParam("docid")
			},
			expected: &URL{
				baseURL: "https://video.google.co.uk:80/videoplay",
				queryParams: map[string][]string{
					"hl": {"en"},
				},
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewURL(tt.input)
			if err != nil {
				t.Fatalf("bad test (%v)", tt.name)
			}

			// apply modifications from test scenario
			tt.modifier(output)

			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}
