package zendesksupport

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func TestLookupPaginationType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		module common.ModuleID
		object string
		want   string
	}{
		{
			module: ModuleTicketing,
			object: "tickets",
			want:   "cursor",
		},
		{
			module: ModuleTicketing,
			object: "workspaces",
			want:   "offset",
		},
		{
			module: ModuleHelpCenter,
			object: "articles",
			want:   "offset",
		},
		{
			module: ModuleTicketing,
			object: "macros",
			want:   "cursor",
		},
		{
			module: ModuleHelpCenter,
			object: "community_posts",
			want:   "offset",
		},
	}

	for _, test := range tests {
		t.Run(test.object, func(t *testing.T) {
			got, ok := metadata.Schemas.LookupPaginationType(test.module, test.object)
			if test.want != got || !ok {
				t.Errorf("LookupPaginationType(%s) = %v, want %v", test.object, got, test.want)
			}
		})
	}
}
