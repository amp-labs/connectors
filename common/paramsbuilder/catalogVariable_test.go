package paramsbuilder

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common/substitutions"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

func TestNewCatalogVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    substitutions.Registry[string]
		expected []catalogreplacer.CatalogVariable
	}{
		{
			name: "Unknown substitutions is translated to Variables",
			input: substitutions.Registry[string]{
				"subdomain": "www",
			},
			expected: []catalogreplacer.CatalogVariable{
				createCustomVariable("subdomain", "www"),
			},
		},
		{
			name: "Workspace Variable is captured",
			input: substitutions.Registry[string]{
				"workspace": "office",
			},
			expected: []catalogreplacer.CatalogVariable{
				&Workspace{Name: "office"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := NewCatalogVariables(tt.input)
			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

func TestNewCatalogSubstitutionRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []catalogreplacer.CatalogVariable
		expected substitutions.Registry[string]
	}{
		{
			name:     "No variables - no substitutions",
			input:    []catalogreplacer.CatalogVariable{},
			expected: substitutions.Registry[string]{},
		},
		{
			name: "Workspace is translated into substitution",
			input: []catalogreplacer.CatalogVariable{
				&Workspace{Name: "cool organization"},
			},
			expected: substitutions.Registry[string]{
				catalogreplacer.VariableWorkspace: "cool organization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := catalogreplacer.NewCatalogSubstitutionRegistry(tt.input)
			if !reflect.DeepEqual(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

func createCustomVariable(from, to string) *catalogreplacer.CustomCatalogVariable {
	return &catalogreplacer.CustomCatalogVariable{
		Plan: catalogreplacer.SubstitutionPlan{
			From: from,
			To:   to,
		},
	}
}
