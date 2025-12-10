package salesforce

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	records, err := jsonquery.New(node).ArrayRequired("records")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(records)
}

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextRecordsUrl", "")
}

// getSalesforceDataMarshaller returns a marshaller that fills Associations in ReadResultRow for Salesforce.
func getSalesforceDataMarshaller(
	config common.ReadParams,
) func(
	[]map[string]any,
	[]string,
) ([]common.ReadResultRow, error) {
	// This is a common.MarshalFunc.
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		data := make([]common.ReadResultRow, len(records))

		// Go through each record, attach associations (if any) to the record and
		// convert the record to a common.ReadResultRow.
		for idx, record := range records {
			recordMap := common.ToStringMap(record)

			// Extract the ID of the record.
			id, _ := recordMap.GetCaseInsensitive("Id")
			idStr, _ := id.(string)

			data[idx] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
				Raw:    record,
				Id:     idStr,
			}

			associations := extractAssociations(recordMap, config)

			if len(associations) > 0 {
				data[idx].Associations = associations
			}
		}

		return data, nil
	}
}

// extractAssociations extracts associations from a record map.
// There are two types of relationships:
//  1. Parent relationships (e.g., Opportunity -> Account via AccountId): We only have the parent field
//     value (the ID) in the response. We create an association with empty Raw, which triggers the
//     workflow layer to fetch the full associated record using GetRecordsByIds.
//  2. Child relationships (e.g., Account -> Contacts): The associated records come nested in the
//     response, so we can extract them directly.
func extractAssociations(recordMap common.StringMap, config common.ReadParams) map[string][]common.Association {
	associations := make(map[string][]common.Association)

	for _, assocObj := range config.AssociatedObjects {
		var assoc []common.Association
		if isParentRelationship(config.ObjectName, assocObj) {
			assoc = extractParentAssociation(recordMap, config.ObjectName, assocObj)
		} else {
			assoc = extractChildAssociation(recordMap, assocObj)
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
