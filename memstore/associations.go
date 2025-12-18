package memstore

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// expandAssociations orchestrates the expansion of all requested associations for a set of records.
// It returns a map structure: recordID -> associationName -> []Association
//
// The function handles all three types of associations:
//   - Foreign key (many-to-one): Fetches the parent record referenced by a foreign key field
//   - Reverse lookup (one-to-many): Fetches child records that reference this record
//   - Junction (many-to-many): Fetches related records through an intermediate junction object
//
// Parameters:
//   - objectName: The type of object being read (e.g., "contact", "account")
//   - records: The records that were fetched, which may contain foreign key values
//   - requestedAssociations: The list of association names to expand (from ReadParams.AssociatedObjects)
//
// Returns:
//   - A nested map where the first key is the record ID, the second key is the association name,
//     and the value is a slice of Association objects containing the related records
//   - An error if association expansion fails
//
//nolint:cyclop,funlen // Complexity from handling multiple association types
func (c *Connector) expandAssociations(
	objectName string,
	records []map[string]any,
	requestedAssociations []string,
) (map[string]map[string][]common.Association, error) {
	if len(requestedAssociations) == 0 || len(records) == 0 {
		return make(map[string]map[string][]common.Association), nil
	}

	// Get association metadata for this object
	allAssociations := c.storage.GetAssociations()[ObjectName(objectName)]
	if len(allAssociations) == 0 {
		// No associations defined for this object, nothing to expand
		return make(map[string]map[string][]common.Association), nil
	}

	// Build result structure: recordID -> associationName -> []Association
	result := make(map[string]map[string][]common.Association)

	// Get the ID field name for this object to extract record IDs
	idField := c.storage.GetIdFields()[ObjectName(objectName)]
	if idField == "" {
		idField = "id"
	}

	// Process each requested association
	for _, requestedName := range requestedAssociations {
		// Find the field(s) that match this association name
		// An association can be requested by field name or by target object name
		var matchingFields []string

		for fieldName, assoc := range allAssociations {
			// Match by field name or by target object name
			if fieldName == requestedName || assoc.TargetObject == requestedName {
				matchingFields = append(matchingFields, fieldName)
			}
		}

		if len(matchingFields) == 0 {
			// Association not found, skip it (graceful degradation)
			continue
		}

		// Expand each matching field's association
		for _, fieldName := range matchingFields {
			assoc := allAssociations[fieldName]

			var expanded map[string][]common.Association

			var err error

			switch assoc.AssociationType {
			case "foreignKey":
				expanded, err = c.expandForeignKeyAssociation(records, fieldName, assoc, idField)
			case "reverseLookup":
				expanded, err = c.expandReverseLookupAssociation(records, fieldName, assoc, idField)
			case "junction":
				expanded, err = c.expandJunctionAssociation(records, fieldName, assoc, idField)
			default:
				// Unknown association type, skip
				continue
			}

			if err != nil {
				return nil, fmt.Errorf("failed to expand association %s: %w", fieldName, err)
			}

			// Merge the expanded associations into the result
			for recordID, associations := range expanded {
				if result[recordID] == nil {
					result[recordID] = make(map[string][]common.Association)
				}
				// Use field name as the association key
				result[recordID][fieldName] = associations
			}
		}
	}

	return result, nil
}

// expandForeignKeyAssociation handles many-to-one relationships where a field contains
// a foreign key (ID) referencing another record.
//
// For example, if a "contact" record has an "account_id" field, this function fetches
// the corresponding "account" record and returns it as an Association.
//
// Algorithm:
//  1. Extract the foreign key value from each record
//  2. Skip records where the foreign key is null/missing
//  3. Fetch the referenced record from storage
//  4. Create an Association object with the fetched data
//
// Returns: map[recordID][]Association where each record gets 0 or 1 association.
func (c *Connector) expandForeignKeyAssociation(
	records []map[string]any,
	fieldName string,
	assoc *AssociationSchema,
	idField string,
) (map[string][]common.Association, error) {
	result := make(map[string][]common.Association)

	for _, record := range records {
		// Get the record's ID
		recordID, ok := record[idField]
		if !ok {
			continue
		}

		recordIDStr := fmt.Sprintf("%v", recordID)

		// Get the foreign key value (the ID of the referenced record)
		foreignKeyValue, exists := record[fieldName]
		if !exists || foreignKeyValue == nil {
			// No foreign key set, no association to expand
			continue
		}

		foreignKeyID := fmt.Sprintf("%v", foreignKeyValue)

		// Fetch the referenced record from storage
		referencedRecord, err := c.storage.Get(assoc.TargetObject, foreignKeyID)
		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				// Referenced record doesn't exist, skip gracefully
				continue
			}

			return nil, fmt.Errorf("failed to fetch %s record %s: %w", assoc.TargetObject, foreignKeyID, err)
		}

		// Create the association
		association := common.Association{
			ObjectId: foreignKeyID,
			Raw:      referencedRecord,
		}

		result[recordIDStr] = []common.Association{association}
	}

	return result, nil
}

// expandReverseLookupAssociation handles one-to-many relationships where we need to find
// all records in the target object that reference the current record.
//
// For example, if an "account" record wants to find all "contact" records that reference it,
// this function scans all contacts and filters for those with matching "account_id".
//
// Algorithm:
//  1. For each record, get its ID
//  2. Fetch all records from the target object
//  3. Filter to those where the foreign key field matches this record's ID
//  4. Create Association objects for all matches
//
// Returns: map[recordID][]Association where each record can have 0-N associations.
//
//nolint:cyclop // Complexity from record matching and validation logic
func (c *Connector) expandReverseLookupAssociation(
	records []map[string]any,
	_ string, // fieldName is unused but kept for interface consistency
	assoc *AssociationSchema,
	idField string,
) (map[string][]common.Association, error) {
	if assoc.ForeignKeyField == "" {
		return nil, fmt.Errorf("%w: reverseLookup requires ForeignKeyField", ErrInvalidAssociation)
	}

	// Fetch all records from the target object once (optimization)
	targetRecords, err := c.storage.GetAll(assoc.TargetObject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s records: %w", assoc.TargetObject, err)
	}

	// Get the ID field name for the target object
	targetIDField := c.storage.GetIdFields()[ObjectName(assoc.TargetObject)]
	if targetIDField == "" {
		targetIDField = "id"
	}

	result := make(map[string][]common.Association)

	for _, record := range records {
		// Get the record's ID
		recordID, ok := record[idField]
		if !ok {
			continue
		}

		recordIDStr := fmt.Sprintf("%v", recordID)

		// Find all target records that reference this record
		var matchingAssociations []common.Association

		for _, targetRecord := range targetRecords {
			// Check if this target record's foreign key matches our record ID
			foreignKeyValue, exists := targetRecord[assoc.ForeignKeyField]
			if !exists || foreignKeyValue == nil {
				continue
			}

			foreignKeyStr := fmt.Sprintf("%v", foreignKeyValue)
			if foreignKeyStr == recordIDStr {
				// This target record references our record
				targetRecordID, ok := targetRecord[targetIDField]
				if !ok {
					continue
				}

				association := common.Association{
					ObjectId: fmt.Sprintf("%v", targetRecordID),
					Raw:      targetRecord,
				}
				matchingAssociations = append(matchingAssociations, association)
			}
		}

		if len(matchingAssociations) > 0 {
			result[recordIDStr] = matchingAssociations
		}
	}

	return result, nil
}

// expandJunctionAssociation handles many-to-many relationships through a junction table.
//
// For example, if "student" and "course" have a many-to-many relationship through
// a "enrollment" junction table, this function:
//  1. Finds all enrollment records where the student ID matches
//  2. Extracts the course IDs from those enrollments
//  3. Fetches the actual course records
//
// Algorithm:
//  1. For each record, get its ID
//  2. Fetch all junction records
//  3. Filter to junction records where JunctionFromField matches this record's ID
//  4. Extract the target IDs from JunctionToField
//  5. Fetch each target record
//  6. Create Association objects for all matches
//
// Returns: map[recordID][]Association where each record can have 0-N associations.
//
//nolint:cyclop,gocognit,funlen // Complexity from junction table traversal logic
func (c *Connector) expandJunctionAssociation(
	records []map[string]any,
	_ string, // fieldName is unused but kept for interface consistency
	assoc *AssociationSchema,
	idField string,
) (map[string][]common.Association, error) {
	if assoc.JunctionObject == "" || assoc.JunctionFromField == "" || assoc.JunctionToField == "" {
		return nil, fmt.Errorf(
			"%w: junction requires JunctionObject, JunctionFromField, and JunctionToField",
			ErrInvalidAssociation,
		)
	}

	// Fetch all junction records once (optimization)
	junctionRecords, err := c.storage.GetAll(assoc.JunctionObject)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s records: %w", assoc.JunctionObject, err)
	}

	result := make(map[string][]common.Association)

	for _, record := range records {
		// Get the record's ID
		recordID, ok := record[idField]
		if !ok {
			continue
		}

		recordIDStr := fmt.Sprintf("%v", recordID)

		// Find all junction records that reference this record
		var targetIDs []string

		for _, junctionRecord := range junctionRecords {
			// Check if this junction record's "from" field matches our record ID
			fromValue, exists := junctionRecord[assoc.JunctionFromField]
			if !exists || fromValue == nil {
				continue
			}

			fromValueStr := fmt.Sprintf("%v", fromValue)
			if fromValueStr == recordIDStr {
				// This junction record links to our record, extract the target ID
				toValue, exists := junctionRecord[assoc.JunctionToField]
				if exists && toValue != nil {
					targetIDs = append(targetIDs, fmt.Sprintf("%v", toValue))
				}
			}
		}

		// Fetch all target records
		var matchingAssociations []common.Association

		for _, targetID := range targetIDs {
			targetRecord, err := c.storage.Get(assoc.TargetObject, targetID)
			if err != nil {
				if errors.Is(err, ErrRecordNotFound) {
					// Target record doesn't exist, skip gracefully
					continue
				}

				return nil, fmt.Errorf("failed to fetch %s record %s: %w", assoc.TargetObject, targetID, err)
			}

			association := common.Association{
				ObjectId: targetID,
				Raw:      targetRecord,
			}
			matchingAssociations = append(matchingAssociations, association)
		}

		if len(matchingAssociations) > 0 {
			result[recordIDStr] = matchingAssociations
		}
	}

	return result, nil
}

// validateAssociations validates that all foreign key references in a record point to existing records.
// This ensures referential integrity during Write operations.
//
// Only foreign key associations are validated on write:
//   - reverseLookup: Not validated (we don't control what other records reference us)
//   - junction: Not validated here (junction records are separate objects)
//
// Returns an error if any foreign key references a non-existent record.
func (c *Connector) validateAssociations(objectName string, record map[string]any) error {
	// Get association metadata for this object
	associations := c.storage.GetAssociations()[ObjectName(objectName)]
	if len(associations) == 0 {
		// No associations to validate
		return nil
	}

	// Validate each foreign key association
	for fieldName, assoc := range associations {
		// Only validate foreign keys on write
		if assoc.AssociationType != "foreignKey" {
			continue
		}

		// Extract the foreign key value from the record
		foreignKeyValue, exists := record[fieldName]
		if !exists || foreignKeyValue == nil {
			// No foreign key set, nothing to validate (null is allowed unless the field is required)
			continue
		}

		foreignKeyID := fmt.Sprintf("%v", foreignKeyValue)

		// Verify that the referenced record exists
		_, err := c.storage.Get(assoc.TargetObject, foreignKeyID)
		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				return fmt.Errorf(
					"%w: field %s references %s record %s which does not exist",
					ErrInvalidForeignKey,
					fieldName,
					assoc.TargetObject,
					foreignKeyID,
				)
			}

			return fmt.Errorf("failed to validate foreign key %s: %w", fieldName, err)
		}
	}

	return nil
}
