package api3

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSchemaRefComponentName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ref      string
		expected string
	}{
		{
			name:     "OpenAPI 3 component ref",
			ref:      "#/components/schemas/CompanyReference",
			expected: "CompanyReference",
		},
		{
			name:     "Swagger 2 definition ref",
			ref:      "#/definitions/ContactReference",
			expected: "ContactReference",
		},
		{
			name:     "empty ref",
			ref:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := schemaRefComponentName(tt.ref); got != tt.expected {
				t.Errorf("schemaRefComponentName(%q) = %q, want %q", tt.ref, got, tt.expected)
			}
		})
	}
}

func TestExtractPropertySchemaRefName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   *openapi3.SchemaRef
		expected string
	}{
		{
			name: "object property referencing a component",
			schema: &openapi3.SchemaRef{
				Ref:   "#/components/schemas/CompanyReference",
				Value: &openapi3.Schema{},
			},
			expected: "CompanyReference",
		},
		{
			name: "array property with referenced items",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Items: &openapi3.SchemaRef{
						Ref:   "#/components/schemas/ContactTypeReference",
						Value: &openapi3.Schema{},
					},
				},
			},
			expected: "ContactTypeReference",
		},
		{
			name: "inlined property schema",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{},
			},
			expected: "",
		},
		{
			name: "object property wrapping a reference in allOf",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					AllOf: openapi3.SchemaRefs{{
						Ref:   "#/components/schemas/CompanyReference",
						Value: &openapi3.Schema{},
					}},
				},
			},
			expected: "CompanyReference",
		},
		{
			name: "array property declared via named wrapper schema",
			schema: &openapi3.SchemaRef{
				Ref: "#/components/schemas/CompanyReferenceList",
				Value: &openapi3.Schema{
					Items: &openapi3.SchemaRef{
						Ref:   "#/components/schemas/CompanyReference",
						Value: &openapi3.Schema{},
					},
				},
			},
			expected: "CompanyReference",
		},
		{
			name: "ambiguous composition of multiple distinct references",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					OneOf: openapi3.SchemaRefs{{
						Ref:   "#/components/schemas/CompanyReference",
						Value: &openapi3.Schema{},
					}, {
						Ref:   "#/components/schemas/ContactReference",
						Value: &openapi3.Schema{},
					}},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractPropertySchemaRefName(tt.schema); got != tt.expected {
				t.Errorf("extractPropertySchemaRefName() = %q, want %q", got, tt.expected)
			}
		})
	}
}
