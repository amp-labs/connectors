package associations

import "strings"

// This helps us identify if we can use a SOQL subquery to get an associated object, since SOQL subqueries
// only work for child objects.
func getParentFieldMap() map[string]map[string]string {
	return map[string]map[string]string{
		"opportunity": {
			"accounts": "AccountId",
		},
	}
}

func isParentRelationship(objectName, associatedObject string) bool {
	parentFieldMap := getParentFieldMap()

	objMap, ok := parentFieldMap[strings.ToLower(objectName)]
	if !ok {
		return false
	}

	_, ok = objMap[strings.ToLower(associatedObject)]

	return ok
}

func getParentFieldName(objectName, associatedObject string) string {
	parentFieldMap := getParentFieldMap()

	objMap, ok := parentFieldMap[strings.ToLower(objectName)]
	if !ok {
		return ""
	}

	return objMap[strings.ToLower(associatedObject)]
}

// containsField checks if a field exists in the fields list (case-insensitive).
// e.g. containsField(["Id", "Name", "AccountId"], "accountid") -> true.
func containsField(fields []string, fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	for _, field := range fields {
		if strings.ToLower(field) == fieldLower {
			return true
		}
	}

	return false
}

// junctionRelationshipMapping defines the relationship name and related ID field for junction relationships.
type junctionRelationshipMapping struct {
	RelationshipName string // e.g., "OpportunityContactRoles"
	RelatedIdField   string // e.g., "ContactId"
}

// getJunctionRelationshipMap returns the junction relationship map.
// This is used for junction objects where we query a child relationship but extract a related object ID.
// e.g., For Opportunity -> contacts, we query OpportunityContactRoles but extract ContactId.
func getJunctionRelationshipMap() map[string]map[string]junctionRelationshipMapping {
	return map[string]map[string]junctionRelationshipMapping{
		"opportunity": {
			"contacts": {
				RelationshipName: "OpportunityContactRoles",
				RelatedIdField:   "ContactId",
			},
		},
	}
}

// getJunctionRelationship returns the relationship name and related ID field for junction relationships.
// Returns ok=false if this is not a junction relationship.
func getJunctionRelationship(objectName, associatedObject string) (relationshipName, relatedIdField string, ok bool) {
	junctionRelationshipMap := getJunctionRelationshipMap()

	objMap, found := junctionRelationshipMap[strings.ToLower(objectName)]

	if !found {
		return "", "", false
	}

	mapping, found := objMap[strings.ToLower(associatedObject)]
	if !found {
		return "", "", false
	}

	return mapping.RelationshipName, mapping.RelatedIdField, true
}
