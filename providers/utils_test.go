// nolint:ireturn
package providers

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common/paramsbuilder"
)

var (
	testCatalog CatalogType = map[string]ProviderInfo{ // nolint:gochecknoglobals
		"test": {
			AuthType:    Oauth2,
			Name:        "test",
			BaseURL:     "https://{{.workspace}}.test.com",
			DisplayName: "Super Test",
		},
	}
	customTestCatalogOption = []CatalogOption{ // nolint:gochecknoglobals
		func(params *catalogParams) {
			params.catalog = &CatalogWrapper{
				Catalog:   testCatalog,
				Timestamp: time.Now().Format(time.RFC3339),
			}
		},
	}
)

func TestNewCustomCatalog(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name         string
		input        []CatalogOption
		expected     CatalogType
		expectedErrs []error
	}{
		{
			name: "Removing catalog is not allowed",
			input: []CatalogOption{
				func(params *catalogParams) {
					params.catalog = nil
				},
			},
			expected:     nil,
			expectedErrs: []error{ErrCatalogNotFound},
		},
		{
			name:         "Custom catalog can be set",
			input:        customTestCatalogOption,
			expected:     testCatalog,
			expectedErrs: nil,
		},
		{
			name:         "Builtin catalog is used by default",
			input:        []CatalogOption{},
			expected:     catalog,
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewCustomCatalog(tt.input...).catalog()
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if output != nil {
				if !reflect.DeepEqual(output.Catalog, tt.expected) {
					t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
				}
			}
		})
	}
}

func TestReadInfo(t *testing.T) { // nolint:funlen
	t.Parallel()

	type inType struct {
		options  []CatalogOption
		provider Provider
		vars     []paramsbuilder.CatalogVariable
	}

	tests := []struct {
		name         string
		input        inType
		expected     *ProviderInfo
		expectedErrs []error
	}{
		{
			name: "Returns missing provider error",
			input: inType{
				options:  customTestCatalogOption,
				provider: "nobody knows",
				vars:     nil,
			},
			expected:     nil,
			expectedErrs: []error{ErrProviderNotFound},
		},
		{
			name: "Works without substitution",
			input: inType{
				options:  customTestCatalogOption,
				provider: "test",
				vars:     nil,
			},
			expected: &ProviderInfo{
				AuthType:    Oauth2,
				Name:        "test",
				BaseURL:     "https://{{.workspace}}.test.com",
				DisplayName: "Super Test",
			},
			expectedErrs: nil,
		},
		{
			name: "Works with substitution",
			input: inType{
				options:  customTestCatalogOption,
				provider: "test",
				vars: []paramsbuilder.CatalogVariable{
					&paramsbuilder.Workspace{Name: "europe"},
				},
			},
			expected: &ProviderInfo{
				AuthType:    Oauth2,
				Name:        "test",
				BaseURL:     "https://europe.test.com",
				DisplayName: "Super Test",
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output, err := NewCustomCatalog(tt.input.options...).
				ReadInfo(tt.input.provider, tt.input.vars...)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}
