package dynamicscrm

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
			input:    createURLWithQuery("$select", []string{"cat"}),
			expected: "https://test?$select=cat",
		},
		{
			name:     "Many parameters",
			input:    createURLWithQuery("$select", []string{"cat", "dog", "parrot", "hamster"}),
			expected: "https://test?$select=cat,dog,parrot,hamster",
		},
		{
			name:     "OData parameters with @ symbol",
			input:    createURLWithQuery("$select", []string{"cat", "@odata.dog", "parrot"}),
			expected: "https://test?$select=cat,@odata.dog,parrot",
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

	link, err := constructURL("https://test")
	if err != nil {
		panic(fmt.Errorf("test is incorrect %w", err))
	}

	if len(key) != 0 {
		link.WithQueryParam(key, value)
	}

	return link
}
