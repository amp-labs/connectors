package suiteql

import (
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
)

// TestMakeSuiteQLBody_DateFilter checks that incremental reads filter on the right column, and
// that tables without lastModifiedDate are read in full instead of sending a query SuiteQL rejects.
func TestMakeSuiteQLBody_DateFilter(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		params common.ReadParams
		want   string
	}{
		{
			name:   "full read has no filter",
			params: common.ReadParams{ObjectName: "customer"},
			want:   "SELECT * FROM customer",
		},
		{
			name:   "incremental read filters on lastModifiedDate",
			params: common.ReadParams{ObjectName: "customer", Since: since, Until: until},
			want: "SELECT * FROM customer WHERE " +
				"lastModifiedDate >= TO_TIMESTAMP('2026-10-01 00:00:00.000000000', 'YYYY-MM-DD HH24:MI:SSxFF') AND " +
				"lastModifiedDate <= TO_TIMESTAMP('2026-11-01 00:00:00.000000000', 'YYYY-MM-DD HH24:MI:SSxFF')",
		},
		{
			name:   "transactionline filters on lineLastModifiedDate",
			params: common.ReadParams{ObjectName: "transactionline", Since: since},
			want: "SELECT * FROM transactionline WHERE " +
				"lineLastModifiedDate >= TO_TIMESTAMP('2026-10-01 00:00:00.000000000', 'YYYY-MM-DD HH24:MI:SSxFF')",
		},
		{
			name:   "nexus has no lastModifiedDate, so it is read in full",
			params: common.ReadParams{ObjectName: "nexus", Since: since, Until: until},
			want:   "SELECT * FROM nexus",
		},
		{
			name:   "table names match case-insensitively",
			params: common.ReadParams{ObjectName: "TaxGroupTaxItem", Since: since},
			want:   "SELECT * FROM TaxGroupTaxItem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := makeSuiteQLBody(tt.params).Query; got != tt.want {
				t.Errorf("makeSuiteQLBody() query =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
