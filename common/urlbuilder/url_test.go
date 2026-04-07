package urlbuilder

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
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

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
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

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
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

func TestAddPath(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "No path",
			input:    nil,
			expected: "",
		},
		{
			name:     "Parts without slashes",
			input:    []string{"carts", "personal", "products"},
			expected: "/carts/personal/products",
		},
		{
			name:     "Slashes at the beginning/end of URI part",
			input:    []string{"carts/", "personal/", "/products"},
			expected: "/carts/personal/products",
		},
		{
			name:     "Double and triple slashes",
			input:    []string{"carts///", "/personal//", "/products"},
			expected: "/carts/personal/products",
		},
		{
			name:     "Empty URI parts are ignored",
			input:    []string{"wishlists", "", "//house", "", "items"},
			expected: "/wishlists/house/items",
		},
		{
			name:     "Slashes as URI parts are ignored",
			input:    []string{"coupons", "/", "/", "redeem"},
			expected: "/coupons/redeem",
		},
		{
			name:     "Trailing slash is missing",
			input:    []string{"search", "/"},
			expected: "/search",
		},
		{
			name:     "Trailing slashes are missing",
			input:    []string{"search", "///"},
			expected: "/search",
		},
		{
			name:     "Path with query params",
			input:    []string{"search", "results?q=hello&v=1"},
			expected: "/search/results?q=hello&v=1",
		},
		{
			name:     "Multiple path segments with query params",
			input:    []string{"search?a=1", "results?b=2"},
			expected: "/search/results?a=1&b=2",
		},
		{
			name:     "Query params with multiple values",
			input:    []string{"search?a=1&a=2", "results?b=3"},
			expected: "/search/results?a=1&a=2&b=3",
		},
		{
			name:     "Path with encoded query params",
			input:    []string{"search", "results?q=hello%20world"},
			expected: "/search/results?q=hello+world",
		},
		{
			name:     "Last path segment with identical query param takes precedence",
			input:    []string{"search?a=25", "results?a=33"},
			expected: "/search/results?a=33",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			base := "https://google.com"
			// Deliberately pass extra slashes which should have no effect
			u, err := New(base + "////")
			if err != nil {
				t.Fatalf("bad test (%v)", tt.name)
			}

			// apply modifications from test scenario
			fullURL := u.AddPath(tt.input...).String()
			// We are testing only the path from root.
			path, _ := strings.CutPrefix(fullURL, base)

			// Query parameters can be in any order, so we need to compare them carefully
			// if they are present in expected.
			if strings.Contains(tt.expected, "?") {
				expectedURL, _ := url.Parse("https://google.com" + tt.expected)
				actualURL, _ := url.Parse(fullURL)

				if expectedURL.Path != actualURL.Path {
					t.Fatalf("%s: expected path: (%v), got: (%v)", tt.name, expectedURL.Path, actualURL.Path)
				}

				if !reflect.DeepEqual(expectedURL.Query(), actualURL.Query()) {
					t.Fatalf("%s: expected query: (%v), got: (%v)", tt.name, expectedURL.Query(), actualURL.Query())
				}
			} else {
				if !reflect.DeepEqual(path, tt.expected) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, path)
				}
			}
		})
	}
}

func TestFromRawURL(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with path",
			input:    "http://video.google.co.uk:80/videoplay",
			expected: "http://video.google.co.uk:80/videoplay",
		},
		{
			name:     "Trailing slash is preserved",
			input:    "http://video.google.co.uk:80/",
			expected: "http://video.google.co.uk:80/",
		},
		{
			name:     "URL with query parameters",
			input:    "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
		},
		{
			name:     "Fragment identifier",
			input:    "https://example.com/page#section",
			expected: "https://example.com/page#section",
		},
		{
			name:     "Spaces in query params are encoded with plus sign (+)",
			input:    "https://example.com/data?id=123&info=Hello%20World",
			expected: "https://example.com/data?id=123&info=Hello+World",
		},
		{
			name:     "IP address instead of hostname",
			input:    "http://192.168.1.1/admin",
			expected: "http://192.168.1.1/admin",
		},
		{
			name:     "WebSocket protocol",
			input:    "wss://example.com/socket",
			expected: "wss://example.com/socket",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endpoint, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("%s: is an invalid test, check input", tt.name)
			}

			outputURL, err := FromRawURL(endpoint)
			output := outputURL.String()
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, err)
		})
	}
}

func TestEquality(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name  string
		a     string
		b     string
		equal bool
	}{
		{
			name:  "Query params in different order",
			a:     "https://example.com/path?a=1&b=2",
			b:     "https://example.com/path?b=2&a=1",
			equal: true,
		},
		{
			name:  "Repeated query params, same values different order",
			a:     "https://example.com/path?a=1&a=2",
			b:     "https://example.com/path?a=2&a=1",
			equal: true,
		},
		{
			name:  "Encoded vs unencoded space in query",
			a:     "https://example.com/path?q=hello+world",
			b:     "https://example.com/path?q=hello%20world",
			equal: true,
		},
		{
			name:  "Empty query vs missing query",
			a:     "https://example.com/path?",
			b:     "https://example.com/path",
			equal: true,
		},
		{
			name:  "Mixed order and encoding",
			a:     "https://example.com/path?x=1&y=a%2Fb",
			b:     "https://example.com/path?y=a%2Fb&x=1",
			equal: true,
		},
		{
			name:  "Case-insensitive host",
			a:     "https://EXAMPLE.com/path?a=1",
			b:     "https://example.com/path?a=1",
			equal: true,
		},
		{
			name:  "No query vs query with just question mark",
			a:     "https://example.com/path",
			b:     "https://example.com/path?",
			equal: true,
		},
		{
			name:  "Same key repeated multiple times in different order",
			a:     "https://example.com/path?x=1&x=2",
			b:     "https://example.com/path?x=2&x=1",
			equal: true,
		},
		{
			name:  "Same URL, different query param order, same fragment",
			a:     "https://example.com/path?a=1&b=2#section1",
			b:     "https://example.com/path?b=2&a=1#section1",
			equal: true,
		},
		{
			name:  "Symbol encoding",
			a:     "https://example.com/reset?email=user@example.com",
			b:     "https://example.com/reset?email=user%40example.com",
			equal: true,
		},
		{
			name:  "Different origin",
			a:     "https://example.com",
			b:     "https://canada.gov",
			equal: false,
		},
		{
			name:  "Different paths",
			a:     "https://example.com/customers",
			b:     "https://example.com/orders",
			equal: false,
		},
		{
			name:  "Different query param values",
			a:     "https://example.com/orders?customer=Bob",
			b:     "https://example.com/orders?customer=Alice",
			equal: false,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			firstURL, err := New(tt.a)
			if err != nil {
				t.Fatalf("%s: is an invalid test, check urls", tt.name)
			}

			secondURL, err := New(tt.b)
			if err != nil {
				t.Fatalf("%s: is an invalid test, check urls", tt.name)
			}

			output := firstURL.Equals(secondURL)
			testutils.CheckOutput(t, tt.name, tt.equal, output)
		})
	}
}

func TestURLOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTP with port",
			input:    "http://example.com:8080/path?query=1",
			expected: "http://example.com:8080",
		},
		{
			name:     "HTTPS without port",
			input:    "https://example.com/some/path",
			expected: "https://example.com",
		},
		{
			name:     "HTTPS with subdomain",
			input:    "https://sub.example.com/abc",
			expected: "https://sub.example.com",
		},
		{
			name:     "HTTP with IP address",
			input:    "http://127.0.0.1:3000/test",
			expected: "http://127.0.0.1:3000",
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			endpoint, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("%s: is an invalid test, check input", tt.name)
			}

			outputURL, err := FromRawURL(endpoint)
			output := outputURL.Origin()
			testutils.CheckOutputWithError(t, tt.name, tt.expected, nil, output, err)
		})
	}
}

func TestWithUnencodedQueryParam(t *testing.T) { // nolint:funlen
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
				u.WithUnencodedQueryParam("videoId", "%3A45565451%3A")
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en&videoId=%3A45565451%3A",
		},
		{
			name:  "Add list query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithUnencodedQueryParamList("videoId", []string{"%3A45565451%3A", "%3A987568%3A"})
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en" +
				"&videoId=%3A45565451%3A&videoId=%3A987568%3A",
		},
		{
			name:  "Replace query parameter from unencode param to encode params",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
			modifier: func(u *URL) {
				u.WithUnencodedQueryParam("videoId", "%3A55555555%3A")
				u.WithUnencodedQueryParam("videoId", "(69874521)")
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en&videoId=(69874521)",
		},
		{
			name:  "Remove query parameter",
			input: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en&videoId=%3A45565451%3A",
			modifier: func(u *URL) {
				u.RemoveQueryParam("videoId")
			},
			expected: "https://video.google.co.uk:80/videoplay?docid=-7246927612831078230&hl=en",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
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
