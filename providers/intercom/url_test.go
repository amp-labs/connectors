package intercom

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

func TestConstructURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *urlbuilder.URL
		expected string
	}{
		{
			name:     "No params no query string",
			input:    createURLWithQuery("", nil),
			expected: "https://test",
		},
		{
			name:     "One parameter",
			input:    createURLWithQuery("per_page", []string{"25"}),
			expected: "https://test?per_page=25",
		},
		{
			name: "Pagination cursor with equals char",
			input: createURLWithQuery("starting_after", []string{
				"WzE3MTU2OTU2NzkwMDAsIjU3Y2NjMmU2LTEyODctNDEwZC1iMDI3LTVjOGU4NzIzMzU3YyIsMl0=",
			}),
			expected: "https://test?starting_after=WzE3MTU2OTU2NzkwMDAsIjU3Y2NjMmU2LTEyODctNDEwZC1iMDI3LTVjOGU4NzIzMzU3YyIsMl0=",
		},
	}

	for _, tt := range tests {
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := tt.input.String()
			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

func createURLWithQuery(key string, values []string) *urlbuilder.URL {
	value := strings.Join(values, ",")

	url, err := constructURL("https://test")
	if err != nil {
		panic(fmt.Errorf("test is incorrect %w", err))
	}

	if len(key) != 0 {
		url.WithQueryParam(key, value)
	}

	return url
}
