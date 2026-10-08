package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

func TestCatalogVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workspace string
		metadata  map[string]string
		want      map[string]string
	}{
		{
			name:      "workspace only",
			workspace: "acme",
			metadata:  nil,
			want:      map[string]string{"workspace": "acme"},
		},
		{
			name:      "metadata keys are available for substitution",
			workspace: "acme",
			metadata:  map[string]string{"apiDomain": "proxy.example.com"},
			want: map[string]string{
				"workspace": "acme",
				"apiDomain": "proxy.example.com",
			},
		},
		{
			name:      "workspace wins over a metadata entry of the same name",
			workspace: "acme",
			metadata:  map[string]string{"workspace": "stale"},
			want:      map[string]string{"workspace": "acme"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := &parameters{}
			params.WithWorkspace(tt.workspace)
			params.WithMetadata(tt.metadata, nil)

			registry := catalogreplacer.NewCatalogSubstitutionRegistry(catalogVariables(params))

			if len(registry) != len(tt.want) {
				t.Errorf("expected %d substitution variables, got %d: %v",
					len(tt.want), len(registry), registry)
			}

			for key, want := range tt.want {
				got, ok := registry[key]
				if !ok {
					t.Errorf("expected substitution variable %q to be present", key)

					continue
				}

				if got != want {
					t.Errorf("substitution variable %q: want %q, got %q", key, want, got)
				}
			}
		})
	}
}
