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

	associations := collectRecordReferenceAssociations(values, associatedObjects)
	if len(associations) == 0 {
		return nil
	}

	return associations
}

func collectRecordReferenceAssociations(
	values map[string]any,
	associatedObjects []string,
) map[string][]common.Association {
	associations := make(map[string][]common.Association)

	for _, entries := range values {
		appendRecordReferenceEntries(associations, entries, associatedObjects)
	}

	return associations
}

func appendRecordReferenceEntries(
	associations map[string][]common.Association,
	entries any,
	associatedObjects []string,
) {
	entryList, ok := entries.([]any)
	if !ok {
		return
	}

	for _, entry := range entryList {
		targetObject, assoc, ok := recordReferenceAssociation(entry, associatedObjects)
		if !ok {
			continue
		}

		associations[targetObject] = append(associations[targetObject], assoc)
	}
}

func recordReferenceAssociation(
	entry any,
	associatedObjects []string,
) (string, common.Association, bool) {
	ref, ok := entry.(map[string]any)
	if !ok {
		return "", common.Association{}, false
	}

	attrType, _ := ref["attribute_type"].(string)
	if attrType != recordReferenceAttributeType {
		return "", common.Association{}, false
	}

	if activeUntil, ok := ref["active_until"]; ok && activeUntil != nil {
		return "", common.Association{}, false
	}

	targetObject, _ := ref["target_object"].(string)
	if targetObject == "" || !slices.Contains(associatedObjects, targetObject) {
		return "", common.Association{}, false
	}

	targetRecordID, _ := ref["target_record_id"].(string)
	if targetRecordID == "" {
		return "", common.Association{}, false
	}

	return targetObject, common.Association{ObjectId: targetRecordID}, true
}
