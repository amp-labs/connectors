package acculynx

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestAttachFinancialAssociations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   common.ReadParams
		parentID string
		raw      map[string]any
		expected map[string][]common.Association
	}{
		{
			name:     "uses the fan-out parent job id",
			params:   common.ReadParams{ObjectName: "jobs/financials", AssociatedObjects: []string{"jobs"}},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: map[string][]common.Association{"jobs": {{ObjectId: "job-parent"}}},
		},
		{
			name:     "attaches nothing without a parent job id",
			params:   common.ReadParams{ObjectName: "jobs/financials", AssociatedObjects: []string{"jobs"}},
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
		{
			name:     "attaches nothing when the association is not requested",
			params:   common.ReadParams{ObjectName: "jobs/financials"},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
		{
			name:     "ignores other objects",
			params:   common.ReadParams{ObjectName: "jobs/invoices", AssociatedObjects: []string{"jobs"}},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rows := []common.ReadResultRow{{Raw: tt.raw}}

			attachFinancialAssociations(tt.params, tt.parentID, rows)

			assert.DeepEqual(t, rows[0].Associations, tt.expected)
		})
	}
}
