package zendesksupport

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
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
			module: common.ModuleID(providers.ModuleZendeskTicketing),
			object: "tickets",
			want:   "cursor",
		},
		{
			module: common.ModuleID(providers.ModuleZendeskTicketing),
			object: "workspaces",
			want:   "offset",
		},
		{
			module: common.ModuleID(providers.ModuleZendeskHelpCenter),
			object: "articles",
			want:   "offset",
		},
		{
			module: common.ModuleID(providers.ModuleZendeskTicketing),
			object: "macros",
			want:   "cursor",
		},
		{
			module: common.ModuleID(providers.ModuleZendeskHelpCenter),
			object: "community_posts",
			want:   "offset",
		},
	}

	for _, test := range tests {
		t.Run(test.object, func(t *testing.T) {
			t.Parallel()

			got := metadata.Schemas.LookupPaginationType(test.module, test.object)
			if test.want != got {
				t.Errorf("LookupPaginationType(%s) = %v, want %v", test.object, got, test.want)
			}
		})
	}
}
