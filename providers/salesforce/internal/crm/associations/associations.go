package associations

import (
	"github.com/amp-labs/connectors/common"
)

// ExtractAssociations extracts associations from a record map.
// There are three types of relationships:
//  1. Parent relationships (e.g., Opportunity -> Account via AccountId): We only have the parent field
//     value (the ID) in the response. We create an association with empty Raw, which triggers the
//     workflow layer to fetch the full associated record using GetRecordsByIds.
//  2. Junction relationships (e.g., Opportunity -> Contact via OpportunityContactRoles): We query a
//     child relationship (OpportunityContactRoles) but extract the related object ID (ContactId) to
//     create associations for the related object (Contact).
//  3. Child relationships (e.g., Account -> Contacts): The associated records come nested in the
//     response, so we can extract them directly.
func ExtractAssociations(recordMap common.StringMap, config common.ReadParams) map[string][]common.Association {
	associations := make(map[string][]common.Association)

	for _, assocObj := range config.AssociatedObjects {
		var assoc []common.Association
		if isParentRelationship(config.ObjectName, assocObj) {
			assoc = extractParentAssociation(recordMap, config.ObjectName, assocObj)
		} else {
			relationshipName, relatedIdField, isJunction := getJunctionRelationship(config.ObjectName, assocObj)
			if isJunction {
				assoc = extractJunctionAssociation(recordMap, relationshipName, relatedIdField)
			} else {
				assoc = extractChildAssociation(recordMap, assocObj)
			}
		}

		if len(assoc) > 0 {
			associations[assocObj] = assoc
		}
	}

	return associations
}

// extractParentAssociation extracts a parent relationship association.
func extractParentAssociation(recordMap common.StringMap, objectName, assocObj string) []common.Association {
	parentField := getParentFieldName(objectName, assocObj)
	parentValue, found := recordMap.GetCaseInsensitive(parentField)

	if !found {
		return nil
	}

	idStr, isString := parentValue.(string)
	if !isString || idStr == "" {
		return nil
	}

	// Create association with empty Raw - workflow layer will fetch it
	return []common.Association{
		{
			ObjectId: idStr,
			Raw:      nil,
		},
	}
}

// extractChildAssociation extracts a child relationship association.
// In Salesforce, the associated object is a key in the record map.
// For example, "Contacts" will be a key in the record map, with an array of associated contacts.
// It appears as a nested object containing a "records" array with the associated data.
// Additionally, it includes metadata such as "done" (a boolean indicating if all records have been fetched)
// and the total record count. There are other keys in the record map, but we only care about the "records" key.
func extractChildAssociation(recordMap common.StringMap, assocObj string) []common.Association {
	key, found := recordMap.GetCaseInsensitive(assocObj)
	if !found {
		return nil
	}

	assocMap, ok := key.(map[string]any) // nolint:varnamelen
	if !ok {
		return nil
	}

	records, ok := assocMap["records"].([]any)
	if !ok {
		return nil
	}

	var result []common.Association

	// For each associated record, extract the ID, convert it into a common.Association and add it to the result.
	for _, record := range records {
		if assocRec, ok := record.(map[string]any); ok {
			id, _ := common.ToStringMap(assocRec).GetCaseInsensitive("Id")
			idStr, _ := id.(string)

			association := common.Association{
				Raw: assocRec,
			}

			if idStr != "" {
				association.ObjectId = idStr
			}

			result = append(result, association)
		}
	}

	return result
}

// extractJunctionAssociation extracts associations from a junction object relationship.
// For example, in order to retrieve Contact records for an Opportunity, we query the OpportunityContactRoles
// relationship. Then, we get the ContactId and create a Contact association.
func extractJunctionAssociation(
	recordMap common.StringMap,
	relationshipName, relatedIdField string,
) []common.Association {
	// Look for the relationship name in the response (e.g., "OpportunityContactRoles")
	key, found := recordMap.GetCaseInsensitive(relationshipName)
	if !found {
		return nil
	}

	assocMap, ok := key.(map[string]any) // nolint:varnamelen
	if !ok {
		return nil
	}

	records, ok := assocMap["records"].([]any)
	if !ok {
		return nil
	}

	var result []common.Association

	// For each junction record, extract the related object ID (e.g., ContactId)
	for _, record := range records {
		if junctionRec, ok := record.(map[string]any); ok {
			recMap := common.ToStringMap(junctionRec)

			// Extract the related object ID (e.g., ContactId from OpportunityContactRole)
			relatedId, found := recMap.GetCaseInsensitive(relatedIdField)
			if !found {
				continue
			}

			idStr, isString := relatedId.(string)
			if !isString || idStr == "" {
				continue
			}

			// Create association with empty Raw - workflow layer will fetch Contact records
			result = append(result, common.Association{
				ObjectId: idStr,
				Raw:      nil, // Will be fetched via GetRecordsByIds
			})
		}
	}

	return result
}
