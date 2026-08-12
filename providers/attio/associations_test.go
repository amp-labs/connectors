package attio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractRecordReferenceAssociationsSkipsWhenNotRequested(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"values": map[string]any{
			"team": []any{
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "people",
					"target_record_id": "891dcbfc-9141-415d-9b2a-2238a6cc012d",
					"active_until":     nil,
				},
			},
		},
	}

	assert.Nil(t, extractRecordReferenceAssociations(raw, nil))
	assert.Nil(t, extractRecordReferenceAssociations(raw, []string{"companies"}))
}

func TestExtractRecordReferenceAssociations(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"values": map[string]any{
			"company": []any{
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "companies",
					"target_record_id": "99a03ff3-0435-47da-95cc-76b2caeb4dab",
					"active_until":     nil,
				},
			},
			"people": []any{
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "people",
					"target_record_id": "891dcbfc-9141-415d-9b2a-2238a6cc012d",
					"active_until":     nil,
				},
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "people",
					"target_record_id": "5e3fb280-007b-495a-a530-9354bde01de1",
					"active_until":     nil,
				},
			},
			"name": []any{
				map[string]any{
					"attribute_type": "text",
					"value":            "Example deal",
					"active_until":     nil,
				},
			},
		},
	}

	associations := extractRecordReferenceAssociations(raw, []string{"companies", "people"})
	require.NotNil(t, associations)

	require.Len(t, associations["companies"], 1)
	assert.Equal(t, "99a03ff3-0435-47da-95cc-76b2caeb4dab", associations["companies"][0].ObjectId)

	require.Len(t, associations["people"], 2)
	assert.Equal(t, "891dcbfc-9141-415d-9b2a-2238a6cc012d", associations["people"][0].ObjectId)
	assert.Equal(t, "5e3fb280-007b-495a-a530-9354bde01de1", associations["people"][1].ObjectId)
}

func TestExtractRecordReferenceAssociationsSkipsInactiveValues(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"values": map[string]any{
			"company": []any{
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "companies",
					"target_record_id": "old-company-id",
					"active_until":     "2024-01-01T00:00:00.000000000Z",
				},
				map[string]any{
					"attribute_type":   "record-reference",
					"target_object":    "companies",
					"target_record_id": "current-company-id",
					"active_until":     nil,
				},
			},
		},
	}

	associations := extractRecordReferenceAssociations(raw, []string{"companies"})
	require.NotNil(t, associations)
	require.Len(t, associations["companies"], 1)
	assert.Equal(t, "current-company-id", associations["companies"][0].ObjectId)
}
