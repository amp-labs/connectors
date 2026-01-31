package associations

import "github.com/amp-labs/connectors/common"

// FieldsForSelectQueryRead adds fields for associated objects to the fields list.
func FieldsForSelectQueryRead(params *common.ReadParams) []string {
	fields := params.Fields.List()

	if params.AssociatedObjects == nil {
		return fields
	}

	for _, obj := range params.AssociatedObjects {
		fields = addFieldForAssociation(fields, params.ObjectName, obj)
	}

	return fields
}

// addFieldForAssociation adds a field or subquery for an associated object.
func addFieldForAssociation(fields []string, objectName, assocObj string) []string {
	// Some objects cannot be queried using a subquery, such as when the associated object is a parent object.
	// In that case, we fetch the associated object's ID as a field, and fetch the full object in the q
	if isParentRelationship(objectName, assocObj) {
		parentField := getParentFieldName(objectName, assocObj)
		if parentField != "" && !containsField(fields, parentField) {
			fields = append(fields, parentField)
		}

		return fields
	}

	// Check for junction relationship (child relationship that maps to a different object)
	relationshipName, _, isJunction := getJunctionRelationship(objectName, assocObj)
	if isJunction {
		// Use the mapped relationship name for SOQL subquery
		// e.g., (SELECT FIELDS(STANDARD) FROM OpportunityContactRoles)
		fields = append(fields, "(SELECT FIELDS(STANDARD) FROM "+relationshipName+")")

		return fields
	}

	// Standard child relationship
	// Generates subqueries like: (SELECT FIELDS(STANDARD) FROM Contacts)
	// Just standard fields for now, because salesforce errors out > 200 fields on an object.
	// Source: https://www.infallibletechie.com/2023/04/parent-child-records-in-salesforce-soql-using-rest-api.html
	fields = append(fields, "(SELECT FIELDS(STANDARD) FROM "+assocObj+")")

	return fields
}
