package zendesksupport

import (
	"testing"

	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func TestLookupPaginationType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		object string
		want   string
	}{
		{
			object: "tickets",
			want:   "cursor",
		},
		{
			object: "workspaces",
			want:   "offset",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := metadata.Schemas.LookupPaginationType(ModuleTicketing, test.object)
			if !ok {
				t.Errorf("LookupPaginationType(%s) = %v, want %v", test.object, got, test.want)
			}
		})
	}
}
