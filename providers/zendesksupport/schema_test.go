package zendesksupport

import (
	"testing"

	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

func TestLookupPaginationType(t *testing.T) {
	t.Parallel()

	tests := []struct {
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
		{
			object: "articles",
			want:   "offset",
		},
		{
			object: "macros",
			want:   "cursor",
		},
		{
			object: "community_posts",
			want:   "offset",
		},
	}

	for _, test := range tests {
		t.Run(test.object, func(t *testing.T) {
			t.Parallel()

			got := metadata.Schemas.LookupPaginationType(test.object)
			if test.want != got {
				t.Errorf("LookupPaginationType(%s) = %v, want %v", test.object, got, test.want)
			}
		})
	}
}
