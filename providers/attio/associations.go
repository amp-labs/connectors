package attio

import (
	"slices"

	"github.com/amp-labs/connectors/common"
)

const recordReferenceAttributeType = "record-reference"

// extractRecordReferenceAssociations builds Associations from record-reference
// attribute values embedded in an Attio record's values map.
func extractRecordReferenceAssociations(
	raw map[string]any,
	associatedObjects []string,
) map[string][]common.Association {
	if len(associatedObjects) == 0 {
		return nil
	}

	values, ok := raw["values"].(map[string]any)
	if !ok || len(values) == 0 {
		return nil
	}

	associations := make(map[string][]common.Association)

	for _, entries := range values {
		entryList, ok := entries.([]any)
		if !ok {
			continue
		}

		for _, entry := range entryList {
			ref, ok := entry.(map[string]any)
			if !ok {
				continue
			}

			attrType, _ := ref["attribute_type"].(string)
			if attrType != recordReferenceAttributeType {
				continue
			}

			if activeUntil, ok := ref["active_until"]; ok && activeUntil != nil {
				continue
			}

			targetObject, _ := ref["target_object"].(string)
			if targetObject == "" || !slices.Contains(associatedObjects, targetObject) {
				continue
			}

			targetRecordID, _ := ref["target_record_id"].(string)
			if targetRecordID == "" {
				continue
			}

			associations[targetObject] = append(associations[targetObject], common.Association{
				ObjectId: targetRecordID,
			})
		}
	}

	if len(associations) == 0 {
		return nil
	}

	return associations
}
